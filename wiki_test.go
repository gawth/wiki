package main

import "testing"

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
		res := ParseQueryResults([]string{td.source})
		if res[0].WikiName != td.res.WikiName {
			t.Errorf("ParseQueryResult: Failed to extract wiki name %v: %v", res[0].WikiName, td.res.WikiName)
		}
		if res[0].LineNum != td.res.LineNum {
			t.Errorf("ParseQueryResult: Failed to extract line num %v: %v", res[0].LineNum, td.res.LineNum)
		}
		if res[0].Text != td.res.Text {
			t.Errorf("ParseQueryResult: Failed to extract text %v: %v", res[0].Text, td.res.Text)
		}

	}
}
