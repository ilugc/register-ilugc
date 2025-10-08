package register_ilugc

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type RegisterIlugc struct {
	Address string
	Server *http.Server
	CsvFile *os.File
	Csv *csv.Writer
}

func CreateRegisterIlugc(address string) *RegisterIlugc {
	self := &RegisterIlugc{}
	self.Address = ":2203"
	if len(address) > 0 {
		self.Address = address
	}
	return self
}

func (self *RegisterIlugc) Init() error {
	self.Server = &http.Server{
		Addr: self.Address,
	}
	CsvFile, err := os.OpenFile("participants.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644);
	if err != nil {
		G.logger.Println(err)
		return err
	}
	self.CsvFile = CsvFile
	self.Csv = csv.NewWriter(self.CsvFile)
	return nil
}

func (self *RegisterIlugc) Close() {
	if self.Csv != nil {
		self.Csv.Flush()
	}
	if self.CsvFile != nil {
		self.CsvFile.Close()
		self.CsvFile = nil
	}
}

func (self *RegisterIlugc) Run() error {
	defer self.Close()

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		html := `
<html>
  <head>
    <style>
      body,canvas {
          width: 100%;
          height: 100%;
          overflow: hidden;
      }
      #registerform {
          display: flex;
          flex-direction: column;
          justify-content: center;
          align-items: center;
          height: 100%;
      }
      #fieldsdiv {
          display: grid;
          grid-template-columns: 1fr 2fr;
      }
      #submit, #status, #title {
          grid-column: 1 / -1;
      }
      #title, #status {
	  text-align: center;
      }
    </style>
  </head>
  <body>
    <div id="registerform">
      <div id="fieldsdiv">
        <label id="title">ILUGC Monthly Meet Register Form</label>
        <label>Name</label>
        <input id="participant_name" />
        <label>Email</label>
        <input id="participant_email" />
        <label>Mobile</label>
        <input id="participant_mobile" />
        <input id="submit" type="button" value="submit" />
        <label id="status"></label>
      </div>
    </div>
    <script src="/register.js"></script>
  </body>
</html>
`
		response.Write([]byte(html))
	})

	http.HandleFunc("/register.js", func(response http.ResponseWriter, request *http.Request) {
		registerjs := `
let global = {
    submit: document.getElementById("submit"),
    status: document.getElementById("status")
}

let showMessage = function(error) {
    global.submit.focus();
    global.status.innerText = error;
}

global.submit.addEventListener("focusout", (event) => {
    global.status.innerText = "";
});

global.submit.addEventListener("click", (event) => {
    pname = document.getElementById("participant_name");
    if (pname.value.length < 0
	|| /^[A-Za-z ]+$/.test(pname.value) == false) {
	showMessage("invalid name");
	return;
    }
    pemail = document.getElementById("participant_email");
    if (pemail.value.length < 0	||
	/@/.test(pemail.value) == false) {
	showMessage("invalid email");
	return;
    }
    pmobile = document.getElementById("participant_mobile");
    if (pmobile.value.length < 10
	|| pmobile.value.length > 13
	|| /^[+0-9]+$/.test(pmobile.value) == false) {
	showMessage("invalid mobile number");
	return;
    }

    fetch("/register", {
	method: "POST",
	header: {
	    "Content-Type": "application/json"
	},
	body: JSON.stringify({name: pname.value, email: pemail.value, mobile: pmobile.value})
    }).then((response) => {
	if (response.status === 200) {
	    pname.value = '';
	    pemail.value = '';
	    pmobile.value = '';
	    showMessage("registration successful");
	}
	else {
	    showMessage("registration failed, try later");
	}
    }, () => {
	showMessage("registration failed, try later");
    })
})
`
		response.Write([]byte(registerjs))
	})

	http.HandleFunc("/register", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			err := errors.New("request must be POST")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		stat, err := self.CsvFile.Stat()
		if err != nil {
			err := errors.New("cannot stat CsvFile")
			G.logger.Println(err)
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		if stat.Size() > (1024 * 1024 * 512) {
			msg := fmt.Sprint("filesize ", stat.Size(),  " greater than 512MB")
			err := errors.New(msg)
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
		type Participant struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Mobile string `json:"mobile"`
		}
		participant := &Participant{
			Name: body["name"].(string),
			Email: body["email"].(string),
			Mobile: body["mobile"].(string),
		}
		G.logger.Println(participant)
		self.Csv.Write([]string{time.Now().Local().String(), participant.Name, participant.Email, participant.Mobile})
		self.Csv.Flush()
	})
	if err := self.Server.ListenAndServe(); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
