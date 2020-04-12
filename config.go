package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Config object loaded from disk at startup
type Config struct {
	WikiDir       string
	Logfile       string
	CookieKey     string
	KeyLocation   string
	CertPath      string
	KeyPath       string
	HTTPPort      int
	HTTPSPort     int
	UseHTTPS      bool
	EncryptionKey string
}

// getenv returns an env var if it is set or the default passed in
func getenv(key, fallback string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return fallback
	}
	return val
}

// LoadConfig reads in config from file and hydrates to a
// config object
func LoadConfig() (*Config, error) {
	path := "config.json"
	config := Config{
		UseHTTPS:    false,
		WikiDir:     "./wikidir",
		Logfile:     "defaultwiki.log",
		KeyLocation: "./excluded/",
		HTTPPort:    8080,
		HTTPSPort:   8443,
	}
	conf, err := ioutil.ReadFile(path)
	if err == nil {
		log.Printf("Using config file")
		err = json.Unmarshal(conf, &config)
		if err != nil {
			return nil, err
		}
	}

	config.HTTPPort, _ = strconv.Atoi(getenv("PORT", strconv.Itoa(config.HTTPPort)))
	config.HTTPSPort, _ = strconv.Atoi(getenv("HTTPSPORT", strconv.Itoa(config.HTTPSPort)))
	config.KeyLocation = getenv("KEYLOCATION", config.KeyLocation)
	config.UseHTTPS, _ = strconv.ParseBool(getenv("USEHTTPS", strconv.FormatBool(config.UseHTTPS)))
	config.WikiDir = getenv("WIKIDIR", config.WikiDir)
	config.Logfile = getenv("LOGFILE", config.Logfile)
	config.CookieKey = getenv("COOKIEKEY", config.CookieKey)
	config.EncryptionKey = getenv("ENCRYPTIONKEY", config.EncryptionKey)
	if len(config.CookieKey) == 0 {
		return nil, errors.New("Must set a valid cookie key")
	}
	if len(config.EncryptionKey) == 0 {
		return nil, errors.New("Must set a valid, 32 char Encryption Key")
	}

	// Make sure the path ends with a /
	if config.WikiDir[len(config.WikiDir)-1] != '/' {
		config.WikiDir = config.WikiDir + "/"
	}

	if len(config.EncryptionKey) != 32 {
		return nil, errors.New("Need to set EncryptionKey to be 32 char string")
	}
	j, err := json.MarshalIndent(config, "", "   ")
	if err != nil {
		return nil, err
	}
	fmt.Print(string(j))

	return &config, nil

}

// LoadCookieKey gets the secret key that will be used for
// encrypting cookies
func (c *Config) LoadCookieKey() {
	if len(c.CookieKey) == 0 {
		res, err := ioutil.ReadFile(c.KeyLocation + "cookiesecret.txt")
		checkErr(err)
		c.CookieKey = string(res)
	}
	return
}
