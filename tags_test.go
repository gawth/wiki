package main

import "testing"

func TestGetWikisForTag(t *testing.T) {
	expected := []string{"wiki1", "wiki2"}
	target := Tag{}

	// Setup
	for _, v := range expected {
		target.AddWiki(v)
	}

	// Act
	actual := target.GetWikisForTag("testtag")

	// Assert
	if len(actual) != len(expected) {
		t.Errorf("TestGetWikisForTag: Returned %v rather than %v", len(actual), len(expected))
		t.FailNow()
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("TestGetWikisForTag: Returned %v rather than %v", actual[i], expected[i])
		}

	}

}

func TestAssociateTagWiki(t *testing.T) {
	target := TagIndex(make(map[string][]Tag))

	target.AssociateTagWiki("wiki", "tag1")

	if len(target["tag1"]) != 1 {
		t.Errorf("TestAssociateTagWiki: Failed to associate wiki to tag")
	}
}
