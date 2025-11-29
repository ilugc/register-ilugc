package register

import (
	"encoding/json"
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	ConfigDetails
	Filename string `json:"filename"`
	AdminPassword string `json:"adminpassword"`
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

func (self *Config) WriteConfigDetails() error {
	if len(self.AdminUsername) > 0 &&
		len(self.AdminPassword) > 0 {
		bcryptpassword, err := getAdminPasswordHash(self.AdminPassword)
		if err != nil {
			G.logger.Println(err)
		} else {
			self.AdminPasswordHash = bcryptpassword
		}
		self.AdminPassword = ""
	}
	if err := self.Db.ConfigDetailsWrite(&self.ConfigDetails); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Config) LoadConfigDetails() error {
	configdetails, err := self.Db.ConfigDetailsRead();
	if err != nil {
		G.logger.Println(err)
		return err
	}
	self.ConfigDetails = *configdetails
	return nil
}

func (self *Config) Load() error {
	self.LoadConfigDetails()
	content, err := os.ReadFile(self.Filename)
	if err == nil {
		if err := json.Unmarshal(content, self); err != nil {
			G.logger.Println(err)
			return err
		}
		if err := self.WriteConfigDetails(); err != nil {
			G.logger.Println(err)
			return err
		}
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

func (self *Config) ComparePassword(passwordbytes []byte) error {
	if err := bcrypt.CompareHashAndPassword(self.AdminPasswordHash, passwordbytes); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
