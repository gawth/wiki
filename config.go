package main

import (
	"encoding/json"
	"io/ioutil"
)

// Config object loaded from disk at startup
type Config struct {
	WikiDir     string
	CookieKey   []byte
	KeyLocation string
}

// LoadConfig reads in config from file and hydrates to a
// config object
func LoadConfig(path string) (*Config, error) {
	conf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(conf, &config)
	if err != nil {
		return nil, err
	}

	// Make sure the path ends with a /
	config.WikiDir = config.WikiDir + "/"

	// Set a default location for secure files not included in
	// git
	if config.KeyLocation == "" {
		config.KeyLocation = "./excluded/"
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
