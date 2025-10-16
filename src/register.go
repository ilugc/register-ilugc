package register

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type RegisterIlugc struct {
	Config *Config
	Server *http.Server
	Db *Db
	Qr *Qr
	AuthToken map[string]string
}

func CreateRegisterIlugc(config *Config) *RegisterIlugc {
	self := &RegisterIlugc{}

	self.Config = config
	if self.Config == nil {
		self.Config = CreateConfig("")
		self.Config.Init()
	}

	self.Server = &http.Server{
		Addr: self.Config.Hostport,
	}

	self.Db = CreateDb()
	self.Qr = CreateQr()
	self.AuthToken = make(map[string]string)
	return self
}

func (self *RegisterIlugc) Init() error {
	if err := self.Db.Init(); err != nil {
		G.logger.Println(err)
		return err
	}
	if err := self.Qr.Init(); err != nil {
		G.logger.Println(err)
		return err
	}

	if len(self.Config.Static) <= 0 {
		staticpath, err := GetStaticPath()
		if err != nil {
			G.logger.Println(err)
			return err
		}
		G.logger.Println("staticpath:", staticpath)
		self.Config.Static = staticpath
	}
	return nil
}

func (self *RegisterIlugc) Close() {
	db, err := self.Db.Db.DB()
	if err != nil {
		G.logger.Println(err)
		return
	}
	if err := db.Close(); err != nil {
		G.logger.Println(err)
	}
}

func (self *RegisterIlugc) IsClosed() (bool, error) {
	if self.Config.StopRegistration == true {
		return true, nil
	}

	count, err := self.Db.Count()
	if err != nil {
		G.logger.Println(err)
		return false, err
	}
	if self.Config.DefaultMax > 0 &&
		count >= self.Config.DefaultMax {
		return true, nil
	}
	return false, nil
}

func (self *RegisterIlugc) CheckAuth(body map[string]any) error {
	if len(self.Config.AdminUsername) > 0 {
		username, ok := body["AdminUsername"]
		if ok == false {
			err := errors.New("AdminUsername not sent")
			return err
		}
		username = username.(string)
		delete(body, "AdminUsername")
		if self.Config.AdminUsername != username {
			err := errors.New("Authendication Failed")
			return err
		}
	}

	if len(self.Config.AdminPassword) > 0 {
		password, ok := body["AdminPassword"]
		if ok == false {
			err := errors.New("AdminPassword not sent")
			return err
		}
		password = password.(string)
		delete(body, "AdminPassword")
		if self.Config.AdminPassword != password {
			err := errors.New("Authendication Failed")
			return err
		}
	}
	return nil
}

func (self *RegisterIlugc) Run() error {
	defer self.Close()

	http.HandleFunc("/{$}", func(response http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.ParseFiles(self.Config.Static + "/index.tmpl",
			self.Config.Static + "/sourcecode.tmpl"))
		if err := tmpl.Execute(response, nil); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/register/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			err := errors.New("request must be POST")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		isclosed, err := self.IsClosed()
		if err != nil  {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		if isclosed {
			err := errors.New("Registration Closed")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := io.ReadAll(request.Body)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		body := make(map[string]any)
		if err := json.Unmarshal(data, &body); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		participant := &Participant{
			RegisteredTime: time.Now().UTC().Format(time.RFC3339),
			Name: body["name"].(string),
			Email: body["email"].(string),
			Mobile: body["mobile"].(string),
			Org: body["org"].(string),
			Place: body["place"].(string),
			QrCode: nil,
		}

		chksum, err := self.Db.Chksum(participant)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		qrbuffer, err := self.Qr.Gen(self.Config.Domain + "/participant/" + chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		participant.QrCode = qrbuffer.Bytes()
		if err := self.Db.Write(participant); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		type RegisterResp struct {
			Chksum string `json:"chksum"`
		}
		respbody, err := json.Marshal(&RegisterResp{Chksum: chksum})
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		response.Write(respbody)
	})

	http.HandleFunc("/isclosed/", func(response http.ResponseWriter, request *http.Request) {
		isclosed, err := self.IsClosed()
		if err != nil  {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		type IsClosedResp struct {
			IsClosed bool `json:"isclosed"`
		}
		body, err := json.Marshal(&IsClosedResp{IsClosed: isclosed})
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		response.Write(body)
	})

	http.HandleFunc("/participant/{chksum}/", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		participant, err := self.Db.Read(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		participantmap := StructToMap(participant)
		
		type ParticipantResp struct {
			ParticipantMap map[string]string
			UnregisterUrl string
			QrCodeUrl string
		}
		participantresp := &ParticipantResp{ParticipantMap: participantmap, UnregisterUrl: "/delete/" + chksum, QrCodeUrl: "/qr/" + chksum}
		tmpl := template.Must(template.ParseFiles(self.Config.Static + "/participant.tmpl",
			self.Config.Static + "/sourcecode.tmpl"))
		if err := tmpl.Execute(response, participantresp); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/delete/{chksum}/", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		err := self.Db.Delete(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		tmpl := template.Must(template.ParseFiles(self.Config.Static + "/unregister.tmpl",
			self.Config.Static + "/sourcecode.tmpl"))
		if err := tmpl.Execute(response, nil); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/qr/{chksum}/", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		participant, err := self.Db.Read(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		response.Header().Set("Content-Type", "image/png")
		response.Write(participant.QrCode)
	})

	http.HandleFunc("/config/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			configmap := StructToMap(self.Config)
			for _, k := range []string{"Filename", "Hostport", "Static", "AdminUsername", "AdminPassword"} {
				delete(configmap, k)
			}

			tmpl := template.Must(template.ParseFiles(self.Config.Static + "/config.tmpl",
				self.Config.Static + "/admin.tmpl",
				self.Config.Static + "/sourcecode.tmpl"))
			if err := tmpl.Execute(response, configmap); err != nil {
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
			}
			return
		}

		data, err := io.ReadAll(request.Body)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		body := make(map[string]any)
		if err := json.Unmarshal(data, &body); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		if err = self.CheckAuth(body); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		
		StructSetFromMap(self.Config, body)
		response.Write([]byte("Config Updated"))
	})

	http.HandleFunc("/csv/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			path := strings.SplitN(request.URL.Path, "/", 4)
			hashdata := path[2]
			if len(hashdata) <= 0 {
				tmpl := template.Must(template.ParseFiles(self.Config.Static + "/csv.tmpl",
					self.Config.Static + "/admin.tmpl",
					self.Config.Static + "/sourcecode.tmpl"))
				if err := tmpl.Execute(response, nil); err != nil {
					G.logger.Println(err)
					http.Error(response, err.Error(), http.StatusBadRequest)
				}
				return
			}
			timestr, ok := self.AuthToken[hashdata]
			if ok == false {
				err := errors.New("No Auth Token")
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			delete(self.AuthToken, hashdata)

			reqtime, err := time.Parse(time.RFC3339Nano, timestr)
			if err != nil {
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			if time.Now().UTC().Sub(reqtime).Seconds() > 30 {
				err := errors.New("Invalid Auth Token")
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			data, err := self.Db.Csv()
			if err != nil {
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			response.Header().Set("Content-Type", "text/csv")
			response.Header().Set("Content-Disposition", "attachment; filename=\"participants.csv\"")
			response.Write(data)
			return
		}

		data, err := io.ReadAll(request.Body)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		body := make(map[string]any)
		if err := json.Unmarshal(data, &body); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		if err = self.CheckAuth(body); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		nowstring := time.Now().UTC().Format(time.RFC3339Nano)
		hash := base64.URLEncoding.EncodeToString([]byte(nowstring))
		self.AuthToken[hash] = nowstring
		type CsvResp struct {
			Hash string `json:"hash"`
		}
		respbody, err := json.Marshal(&CsvResp{Hash: hash})
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		response.Write(respbody)
	})

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, fmt.Sprint(self.Config.Static, request.URL.Path))
	})

	go func() {
		if err := self.Server.ListenAndServe(); err != nil {
			G.logger.Println(err)
			return
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	G.logger.Println("Received Sinal", <-sig);
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := self.Server.Shutdown(ctx); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
