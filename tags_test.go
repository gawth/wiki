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
	actual := target.GetWikisForTag()

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
	target := TagIndex(make(map[string]Tag))

	target.AssociateTagToWiki("wiki", "tag1")

	if target.GetTag("tag1").TagName != "tag1" {
		t.Errorf("TestAssociateTagWiki: Failed to associate wiki to tag")
	}
	target.AssociateTagToWiki("wiki2", "tag1")

	tg := target.GetTag("tag1")
	if tg.TagName != "tag1" {
		t.Errorf("TestAssociateTagWiki: Failed to associate wiki to tag")
	}

	if len(tg.GetWikisForTag()) != 2 {
		t.Errorf("TestAssociateTagWiki: Wrong num of tags stored:%v", len(tg.GetWikisForTag()))
	}
}
