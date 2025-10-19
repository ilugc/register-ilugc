package register

import (
	"encoding/json"
	"errors"
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
	Admin *Admin
	Db *Db
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
	self.Admin = &Admin{}
	self.Db = CreateDb()
	return self
}

func getAdminPasswordHash(adminpassword string) ([]byte, error) {
	if len(adminpassword) <= 0 {
		err := errors.New("Empty AdminPassword")
		G.logger.Println(err)
		return nil, err
	}

	bcryptbytes, err := bcrypt.GenerateFromPassword([]byte(adminpassword), bcrypt.DefaultCost)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}
	return bcryptbytes, nil
}

func (self *Config) LoadAdmin() error {
	if len(self.AdminUsername) > 0 &&
		len(self.AdminPassword) > 0 {
		bcryptpassword, err := getAdminPasswordHash(self.AdminPassword)
		if err != nil {
			G.logger.Println(err)
			return err
		}
		admin := &Admin{AdminUsername: self.AdminUsername, AdminPassword: bcryptpassword}
		if err := self.Db.AdminWrite(admin); err != nil {
			G.logger.Println(err)
			return err
		}
		self.Admin = admin
	} else {
		admin, err := self.Db.AdminRead();
		if err != nil {
			G.logger.Println(err)
			return err
		}
		self.Admin = admin
	}
	self.AdminUsername = ""
	self.AdminPassword = ""
	return nil
}

func (self *Config) Load() error {
	content, err := os.ReadFile(self.Filename)
	if err == nil {
		if err := json.Unmarshal(content, self); err != nil {
			G.logger.Println(err)
		}
	}

	if err := self.LoadAdmin(); err != nil {
		G.logger.Println(err)
	}
	return nil
}

func (self *Config) Init() error {
	if err := self.Db.Init(); err != nil {
		G.logger.Println(err)
		return err
	}

	if err := self.Load(); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
