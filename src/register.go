package register_ilugc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
)

type Participant struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Mobile string `json:"mobile"`
	Org string `json:"org"`
	Place string `json:"place"`
	QrCode []byte `json:"qrcode"`
	Time string `json:"time"`
}

type RegisterIlugc struct {
	Domain string
	Hostport string
	Static string
	Server *http.Server
	Db *Db
	Qr *Qr
}

func CreateRegisterIlugc(hostport string, static string) *RegisterIlugc {
	self := &RegisterIlugc{}

	self.Domain = "register.ilugc.in"
	self.Hostport = ":2203"
	if len(hostport) > 0 {
		self.Hostport = hostport
	}

	self.Static = "static"
	if len(static) > 0 {
		self.Static = static
	}

	self.Server = &http.Server{
		Addr: self.Hostport,
	}

	self.Db = CreateDb()
	self.Qr = CreateQr()
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

func (self *RegisterIlugc) MaxReached() bool {
	return false;
}

func StructToMap(v any) map[string]string {
	vmap := make(map[string]string)
	typeof := reflect.TypeOf(v)
	valueof := reflect.ValueOf(v)
	for index := 0; index < valueof.NumField(); index++ {
		typef := typeof.Field(index)
		valuef := valueof.Field(index)
		switch valuef.Kind() {
		case reflect.String: {
			vmap[typef.Name] = valuef.String()
		}
		}
	}
	return vmap
}

func (self *RegisterIlugc) Run() error {
	defer self.Close()

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		http.ServeFile(response, request, fmt.Sprint(self.Static, request.URL.Path))
	})

	http.HandleFunc("/register", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			err := errors.New("request must be POST")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		if self.MaxReached() == true {
			err := errors.New("Registration Closed. Maximum participants reached, register early for next month meet.")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := io.ReadAll(request.Body)
		if err != nil {
			err := errors.New("failed to read body")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		body := make(map[string]any)
		if err := json.Unmarshal(data, &body); err != nil {
			err := errors.New("failed to parse body")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		participant := &Participant{
			Name: body["name"].(string),
			Email: body["email"].(string),
			Mobile: body["mobile"].(string),
			Org: body["org"].(string),
			Place: body["place"].(string),
			Time: time.Now().UTC().Format(time.RFC3339),
			QrCode: nil,
		}

		chksum, err := self.Db.Chksum(participant)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		qrbuffer, err := self.Qr.Gen(self.Domain + "/participant/" + chksum)
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
			err := errors.New("failed to generate response body")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		response.Write(respbody)
	})

	http.HandleFunc("/max_reached", func(response http.ResponseWriter, request *http.Request) {
		type MaxReached struct {
			MaxReached bool `json:"max_reached"`
		}
		body, err := json.Marshal(&MaxReached{MaxReached: self.MaxReached()})
		if err != nil {
			err := errors.New("failed to generate response body")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		response.Write(body)
	})

	http.HandleFunc("/participant/{chksum}", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		participant, err := self.Db.Read(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		participantmap := StructToMap(*participant)
		
		type ParticipantResp struct {
			ParticipantMap map[string]string
			UnregisterUrl string
			QrCodeUrl string
		}
		participantresp := &ParticipantResp{ParticipantMap: participantmap, UnregisterUrl: "/delete/" + chksum, QrCodeUrl: "/qr/" + chksum}
		tmpl := template.Must(template.ParseFiles(self.Static + "/participant.tmpl"))
		if err := tmpl.Execute(response, participantresp); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/delete/{chksum}", func(response http.ResponseWriter, request *http.Request) {
		chksum := request.PathValue("chksum")
		err := self.Db.Delete(chksum)
		if err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		tmpl := template.Must(template.ParseFiles(self.Static + "/unregister.tmpl"))
		if err := tmpl.Execute(response, nil); err != nil {
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/qr/{chksum}", func(response http.ResponseWriter, request *http.Request) {
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
