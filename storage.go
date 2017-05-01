package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type storage interface {
	storeFile(name string, content []byte) error
	getPublicPages() []string
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

func indexPubPages(path string) []string {

	var results []string

	err := filepath.Walk(path, func(subpath string, info os.FileInfo, _ error) error {
		if !info.IsDir() {
			results = append(results, strings.TrimPrefix(subpath, path))
		}
		return nil
	})
	checkErr(err)

	return results
}

func (fs fileStorage) getPublicPages() []string {
	return indexPubPages(pubDir)
}
