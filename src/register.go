package register

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
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

type AuthToken struct {
	Now time.Time
	Rand string
	FromTime string
}

type Register struct {
	Config *Config
	Server *http.Server
	Db *Db
	Qr *Qr
	AuthToken map[string]*AuthToken
}

func CreateRegister(config *Config) *Register {
	self := &Register{}

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
	self.AuthToken = make(map[string]*AuthToken)
	return self
}

func (self *Register) Init() error {
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

func (self *Register) Close() {
	db, err := self.Db.Db.DB()
	if err != nil {
		G.logger.Println(err)
		return
	}
	if err := db.Close(); err != nil {
		G.logger.Println(err)
	}
}

type ClosedReason int
const (
	Open ClosedReason = iota
	Closed
	MaxReached
)
var ClosedReasonStrings = map[ClosedReason]string{
	Open: "Open",
	Closed: "Registration Closed",
	MaxReached: "Maximum participants reached. Wait for someone to unregister",
}

func (self *Register) IsClosed() (ClosedReason, error) {
	if self.Config.StopRegistration == true {
		return Closed, nil
	}

	count, err := self.Db.ParticipantCount()
	if err != nil {
		G.logger.Println(err)
		return Closed, err
	}
	if self.Config.DefaultMax > 0 &&
		count >= self.Config.DefaultMax {
		return MaxReached, nil
	}
	return Open, nil
}

func (self *Register) CheckAuth(request *http.Request) error {
	username, passwordb64, ok := request.BasicAuth()

	if len(self.Config.AdminUsername) > 0 {
		if ok == false {
			err := errors.New("Invalid Auth")
			G.logger.Println(err)
			return err
		}
		if len(username) <= 0 {
			err := errors.New("Invalid AuthUsername")
			G.logger.Println(err)
			return err
		}

		if self.Config.AdminUsername != username {
			err := errors.New("Invalid AdminUsername")
			G.logger.Println(err)
			return err
		}
	}

	if self.Config.AdminPasswordHash != nil &&
		len(self.Config.AdminPasswordHash) > 0 {
		if ok == false {
			err := errors.New("Invalid Auth")
			G.logger.Println(err)
			return err
		}

		if len(passwordb64) <= 0 {
			err := errors.New("Invalid AdminPassoword")
			G.logger.Println(err)
			return err
		}

		passworddata, err := base64.StdEncoding.DecodeString(passwordb64)
		if err != nil {
			G.logger.Println(err)
			return err
		}
		type PasswordType struct {
			Rand []byte `json:"rand"`
			Diff []int8 `json:"diff"`
		}
		passwordstruct := &PasswordType{}
		if err := json.Unmarshal(passworddata, passwordstruct); err != nil {
			G.logger.Println(err)
			return err
		}

		if len(passwordstruct.Rand) <= 0 ||
			len(passwordstruct.Diff) <= 0 {
			err := errors.New("Invalid AdminPassoword Data")
			G.logger.Println(err)
			return err
		}

		passwordbytes := make([]byte, len(passwordstruct.Diff))
		for index := 0; index < len(passwordstruct.Diff); index++ {
			passwordbytes[index] = byte(int8(passwordstruct.Rand[index]) - passwordstruct.Diff[index])
		}

		if err := self.Config.ComparePassword(passwordbytes); err != nil {
			G.logger.Println(err)
			return err
		}
	}
	return nil
}

func (self *Register) GenAuthToken(fromtime string) (string, error) {
	authtoken := &AuthToken{Now: time.Now().UTC(), Rand: rand.Text(), FromTime: fromtime}
	buffer := bytes.NewBuffer(nil)
	encoder := gob.NewEncoder(buffer)
	if err := encoder.Encode(authtoken); err != nil {
		G.logger.Println(err)
		return "", err
	}
	hash := sha256.Sum256(buffer.Bytes())
	hashstr := hex.EncodeToString(hash[:])
	self.AuthToken[hashstr] = authtoken
	return hashstr, nil
}

func (self *Register) Run() error {
	defer self.Close()

	http.HandleFunc("/{$}", func(response http.ResponseWriter, request *http.Request) {
		tmpl := template.Must(template.ParseFiles(self.Config.Static + "/index.tmpl",
			self.Config.Static + "/sourcecode.tmpl"))
		type IndexResp struct {
			IsClosed bool
			ClosedReasonString string
		}
		closedreason, err := self.IsClosed()
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		indexresp := &IndexResp{IsClosed: closedreason != Open, ClosedReasonString: ClosedReasonStrings[closedreason]}
		if err := tmpl.Execute(response, indexresp); err != nil {
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

		closedreason, err := self.IsClosed()
		if err != nil  {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		if closedreason != Open {
			err := errors.New(ClosedReasonStrings[closedreason])
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

		chksum, err := self.Db.ParticipantChksum(participant)
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
		if err := self.Db.ParticipantWrite(participant); err != nil {
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

	http.HandleFunc("/participant/{chksum}/", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		participant, err := self.Db.ParticipantRead(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		participantmap := StructToMap(participant)
		
		type ParticipantResp struct {
			ParticipantMap map[string]string
			ParticipantUrl string
			QrCodeUrl string
			UnregisterUrl string
		}
		participantresp := &ParticipantResp{ParticipantMap: participantmap,
			ParticipantUrl: "/participant/" + chksum,
			QrCodeUrl: "/qr/" + chksum,
			UnregisterUrl: "/delete/" + chksum}
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
		err := self.Db.ParticipantDelete(chksum)
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
		participant, err := self.Db.ParticipantRead(chksum)
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
		if err := self.CheckAuth(request); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		
		StructSetFromMap(self.Config, body)
		if err := self.Config.WriteConfigDetails(); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
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
				if err := tmpl.Execute(response, time.Now().UTC().Format(time.RFC3339)); err != nil {
					G.logger.Println(err)
					http.Error(response, err.Error(), http.StatusBadRequest)
				}
				return
			}
			authtoken, ok := self.AuthToken[hashdata]
			if ok == false {
				err := errors.New("No Auth Token")
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			delete(self.AuthToken, hashdata)

			if time.Now().UTC().Sub(authtoken.Now).Seconds() > 30 {
				err := errors.New("Invalid Auth Token")
				G.logger.Println(err)
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			data, err := self.Db.ParticipantCsv(authtoken.FromTime)
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

		if err := self.CheckAuth(request); err != nil {
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
		fromtime := ""
		if fromtimestr, exists := body["fromtime"].(string); exists != false {
			fromtime = fromtimestr
		}

		hash, err := self.GenAuthToken(fromtime)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
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
