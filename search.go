package main

import "strings"

// QueryResults is used to hold search results after a wiki search
type QueryResults struct {
	WikiName string
	LineNum  string
	Text     string
}

// ParseQueryResults converts a result string to a query result
func ParseQueryResults(source []string) []QueryResults {
	res := []QueryResults{}
	for _, r := range source {
		sub := strings.Split(r, "\t")
		if len(sub) < 2 {
			res = append(res, QueryResults{"ERROR", "", "Invalid query result"})
			continue
		}
		res = append(res, QueryResults{
			WikiName: sub[0],
			LineNum:  sub[1],
			Text:     strings.Join(sub[2:], "\t"),
		})

	}
	return res
}
