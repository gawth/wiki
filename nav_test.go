package main

import (
	"testing"
	"time"
)

func genTestNav(name string, dir bool) wikiNav {
	someTime := time.Now()
	someName := "test"

	if len(name) > 0 {
		someName = name
	}

	testNav := wikiNav{
		Name:  someName,
		URL:   "/url",
		IsDir: dir,
		ID:    genID("blahh", "more blahh"),
		Mod:   someTime,
	}

	return testNav
}

func TestRecentsDouble(t *testing.T) {

	testData := []wikiNav{genTestNav("1", false), genTestNav("2", false)}

	results := genRecents(testData)

	if len(results) != 2 {
		t.Errorf("expected 2 but got %v navs", len(results))
	}

	if results[0].Name != "2" {
		t.Errorf("sort didn't work, got %v as first item", results[0].Name)
	}

}
func TestRecentsEmptyDir(t *testing.T) {

	testData := []wikiNav{genTestNav("1", true), genTestNav("2", false)}

	results := genRecents(testData)

	if len(results) != 1 {
		t.Errorf("expected 2 but got %v navs", len(results))
	}

	if results[0].Name != "2" {
		t.Errorf("sort didn't work, got %v as first item", results[0].Name)
	}

}
func TestRecentsDir(t *testing.T) {

	testData := []wikiNav{genTestNav("1", true), genTestNav("2", false)}
	testData[0].SubNav = []wikiNav{genTestNav("3", false)}

	results := genRecents(testData)

	if len(results) != 2 {
		t.Errorf("expected 2 but got %v navs", len(results))
	}

	if results[0].Name != "3" {
		t.Errorf("sort didn't work, got %v as first item", results[0].Name)
	}
	if results[1].Name != "2" {
		t.Errorf("sort didn't work, got %v as first item", results[1].Name)
	}

}
