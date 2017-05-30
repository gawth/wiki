package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// Config object loaded from disk at startup
type Config struct {
	WikiDir       string
	Logfile       string
	CookieKey     []byte
	KeyLocation   string
	CertPath      string
	KeyPath       string
	HTTPPort      int
	HTTPSPort     int
	UseHTTPS      bool
	EncryptionKey string
}

// LoadConfig reads in config from file and hydrates to a
// config object
func LoadConfig(path string) (*Config, error) {

	config := Config{}
	config.HTTPPort = 80
	config.HTTPSPort = 443
	config.KeyLocation = "./excluded/"
	config.UseHTTPS = false

	conf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(conf, &config)
	if err != nil {
		return nil, err
	}

	// Make sure the path ends with a /
	config.WikiDir = config.WikiDir + "/"

	if len(config.EncryptionKey) != 32 {
		return nil, errors.New("Need to set EncryptionKey to be 32 char string")
	}

	return &config, nil

}

// LoadCookieKey gets the secret key that will be used for
// encrypting cookies
func (c *Config) LoadCookieKey() []byte {
	res, err := ioutil.ReadFile(c.KeyLocation + "cookiesecret.txt")
	checkErr(err)
	return res
}
