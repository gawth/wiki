package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/thanhpk/randstr"
)

// Config object loaded from disk at startup
type Config struct {
	WikiDir       string
	Logfile       string
	HTTPPort      int
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
		WikiDir:  "./wikidir",
		Logfile:  "defaultwiki.log",
		HTTPPort: 8080,
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
	config.WikiDir = getenv("WIKIDIR", config.WikiDir)
	config.Logfile = getenv("LOGFILE", config.Logfile)
	config.EncryptionKey = getenv("ENCRYPTIONKEY", config.EncryptionKey)
	if len(config.EncryptionKey) == 0 {
		config.EncryptionKey = randstr.String(32)
		fmt.Printf("Generated EncryptionKey '%v' be sure to add to your config", config.EncryptionKey)
	}

	// Make sure the path ends with a /
	if config.WikiDir[len(config.WikiDir)-1] != '/' {
		config.WikiDir = config.WikiDir + "/"
	}

	if len(config.EncryptionKey) != 32 {
		return nil, fmt.Errorf("Need to set EncryptionKey to be 32 char string not %v",
			len(config.EncryptionKey))
	}
	j, err := json.MarshalIndent(config, "", "   ")
	if err != nil {
		return nil, err
	}
	fmt.Print(string(j))

	return &config, nil

}
