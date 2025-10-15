package register_ilugc

import (
	"os"
	"encoding/json"
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
}

func CreateConfig(filename string) *Config {
	self := &Config{}
	self.Filename = "default.config"
	if len(filename) > 0 {
		self.Filename = filename
	}
	self.Domain = "register.ilugc.in"
	self.Hostport = ":2203"
	self.Static = "static"
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

func (self *Config) Init() error {
	return self.Load()
}
