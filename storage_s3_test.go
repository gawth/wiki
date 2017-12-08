// +build s3

package main

import (
	"bytes"
	"testing"
)

func TestS3(t *testing.T) {
	target := newS3Store("wiki-76635528265")
	filename := "folder/testfile"
	mdfilename := filename + ".md"
	testdata := "This is some data2"

	err := target.storeFile(mdfilename, []byte(testdata))
	if err != nil {
		t.Errorf("Failed to create file %v", err)
	}

	list, err := target.listFiles()
	if err != nil {
		t.Errorf("Failed to list files :%v", err)
	}
	found := false
	for _, v := range list.Contents {
		t.Log(*v.Key)
		if *v.Key == mdfilename {
			found = true
		}
	}
	if !found {
		t.Error("Failed to find the test file")
	}

	file, err := target.getFile(mdfilename)
	if err != nil {
		t.Errorf("Failed to read file :%v", err)
	}
	t.Log(file)
	buf := new(bytes.Buffer)
	buf.ReadFrom(file.Body)
	t.Log(buf.String())
	if buf.String() != testdata {
		t.Errorf("Failed to get right file contents :%v", buf.String())
	}

	page := wikiPage{
		basePage: basePage{
			Title: filename,
		},
	}

	_, err = target.getPage(&page)
	if err != nil {
		t.Errorf("Failed to read page :%v", err)
	}

	searchTerm := "fred"
	searchRes := target.searchPages("", searchTerm)
	if len(searchRes) != 0 {
		t.Errorf("Search - incorrect match for :%v, in %v", searchTerm, searchRes)
	}
	searchTerm = testdata
	searchRes = target.searchPages("", searchTerm)
	if len(searchRes) != 1 {
		t.Errorf("Search - failed to find  one match of :%v in %v", searchTerm, searchRes)
	}

}
