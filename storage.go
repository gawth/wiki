package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type storage interface {
	storeFile(name string, content []byte) error
}

type fileStorage struct {
}

func createDir(filename string) error {
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs fileStorage) storeFile(name string, content []byte) error {
	err := createDir(name)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(name, content, 0600)
	if err != nil {
		return err
	}

	return nil
}
