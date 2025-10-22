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
      body {
        font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
        background-color: #f4f4f9;
        margin: 0;
        padding: 0;
      }

      header {
        background-color: black;
        color: white;
        text-align: center;
        padding: 20px;
      }

      header h1 {
        margin: 0;
        font-size: 28px;
      }

      header h2 {
        color: red;
        margin-top: 5px;
        font-size: 18px;
      }

      #registerform {
        display: flex;
        justify-content: center;
        align-items: center;
        height: calc(100vh - 120px);
      }

      #fieldsdiv {
        background: white;
        padding: 30px 40px;
        border-radius: 10px;
        box-shadow: 0px 4px 8px rgba(0,0,0,0.15);
        display: grid;
        gap: 10px;
        grid-template-columns: 1fr 2fr;
      }

      #fieldsdiv label {
        font-weight: 600;
      }

      #fieldsdiv input[type="text"], 
      #fieldsdiv input[type="email"],
      #fieldsdiv input[type="button"] {
        padding: 10px;
        border: 1px solid #ccc;
        border-radius: 5px;
        font-size: 14px;
      }

      #fieldsdiv input[type="button"] {
        background-color: #0078d7;
        color: white;
        border: none;
        cursor: pointer;
        grid-column: 1 / -1;
        transition: background-color 0.2s ease;
      }

      #fieldsdiv input[type="button"]:hover {
        background-color: #005ea3;
      }

      #status {
        grid-column: 1 / -1;
        text-align: center;
        color: green;
        font-weight: bold;
        margin-top: 10px;
      }

      @media (max-width: 600px) {
        #fieldsdiv {
          grid-template-columns: 1fr;
        }
      }
    </style>
  </head>
  <body>
    <header>
      <h1>India Linux User's Group - Chennai(Madras)</h1>
      <h2>Monthly Meet Register Form</h2>
    </header>
    <div id="registerform">
      <div id="fieldsdiv">
        <label>Name</label>
        <input id="participant_name" type="text" />
        <label>Email</label>
        <input id="participant_email" type="email" />
        <label>Mobile</label>
        <input id="participant_mobile" type="text" />
        <label>College/Company</label>
        <input id="participant_org" type="text" />
        <label>Place</label>
        <input id="participant_place" type="text" placeholder="e.g. Velachery" />
        <input id="submit" type="button" value="Submit" />
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
    porg = document.getElementById("participant_org");
    if (porg.value.length < 0
	|| /^[A-Za-z0-9 ]+$/.test(porg.value) == false) {
	showMessage("invalid organization");
	return;
    }
    pplace = document.getElementById("participant_place");
    if (pplace.value.length < 0
	|| /^[A-Za-z ]+$/.test(pplace.value) == false) {
	showMessage("invalid place");
	return;
    }

    fetch("/register", {
	method: "POST",
	header: {
	    "Content-Type": "application/json"
	},
	body: JSON.stringify({name: pname.value, email: pemail.value, mobile: pmobile.value, org: porg.value, place: pplace.value})
    }).then((response) => {
	if (response.status === 200) {
	    pname.value = '';
	    pemail.value = '';
	    pmobile.value = '';
	    porg.value = '';
	    pplace.value = '';
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
			Org string `json:"org"`
			Place string `json:"place"`
		}
		participant := &Participant{
			Name: body["name"].(string),
			Email: body["email"].(string),
			Mobile: body["mobile"].(string),
			Org: body["org"].(string),
			Place: body["place"].(string),
		}
		G.logger.Println(participant)
		self.Csv.Write([]string{time.Now().Local().String(), participant.Name, participant.Email, participant.Mobile, participant.Org, participant.Place})
		self.Csv.Flush()
	})
	if err := self.Server.ListenAndServe(); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
