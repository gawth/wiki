package main

import (
	"fmt"
	"testing"
)

var ParseData = []struct {
	source string
	res    QueryResults
}{
	{"wiki\t12\tsometext", QueryResults{"wiki", "12", "sometext"}},
	{"wiki\t12\tsometext\tfred", QueryResults{"wiki", "12", "sometext\tfred"}},
	{"wiki\t12", QueryResults{"wiki", "12", ""}},
	{"wiki", QueryResults{"ERROR", "", "Invalid query result"}},
}

func TestParseQueryResults(t *testing.T) {
	for _, td := range ParseData {
		res := ParseQueryResults(td.source)
		if res.WikiName != td.res.WikiName {
			t.Error(fmt.Sprintf("ParseQueryResult: Failed to extract wiki name %v: %v", res.WikiName, td.res.WikiName))
		}
		if res.LineNum != td.res.LineNum {
			t.Error(fmt.Sprintf("ParseQueryResult: Failed to extract line num %v: %v", res.LineNum, td.res.LineNum))
		}
		if res.Text != td.res.Text {
			t.Error(fmt.Sprintf("ParseQueryResult: Failed to extract text %v: %v", res.Text, td.res.Text))
		}

	}
}
