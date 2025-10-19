package register

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Filename string `json:"filename"`
	Domain string `json:"domain"`
	Hostport string `json:"hostport"`
	Static string `json:"static"`
	DefaultMax int64 `json:"defaultmax"`
	StopRegistration bool `json:"stopregistration"`
	AdminUsername string `json:"adminusername"`
	AdminPassword string `json:"adminpassword"`
	AdminPasswordBytes []byte `json:"adminpasswordbytes"`
	AdminPasswordHash []byte `json:"adminpasswordhash"`
}

func CreateConfig(filename string) *Config {
	self := &Config{}
	self.Filename = "default.config"
	if len(filename) > 0 {
		self.Filename = filename
	}
	self.Domain = "https://register.ilugc.in"
	self.Hostport = ":2203"
	self.Static = ""
	self.DefaultMax = 0
	self.StopRegistration = false
	return self
}

func (self *Config) Load() error {
	content, err := os.ReadFile(self.Filename)
	if err != nil {
		G.logger.Println(err)
		return err
	}

	if err := json.Unmarshal(content, self); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Config) GetAdminPassword() string {
	if len(self.AdminPassword) <= 0 {
		G.logger.Println("empty AdminPassword")
		return ""
	}

	if len(self.AdminPasswordBytes) <= 0 {
		adminpasswordbytes, err := base64.StdEncoding.DecodeString(self.AdminPassword)
		if err != nil {
			G.logger.Println(err)
			return ""
		}
		self.AdminPasswordBytes = adminpasswordbytes
	}

	if len(self.AdminPasswordHash) <= 0 {
		bcryptbytes, err := bcrypt.GenerateFromPassword(self.AdminPasswordBytes, bcrypt.DefaultCost)
		if err != nil {
			G.logger.Println(err)
			return ""
		}
		self.AdminPasswordHash = bcryptbytes
	}
	return string(self.AdminPasswordBytes)
}

func (self *Config) Init() error {
	return self.Load()
}
