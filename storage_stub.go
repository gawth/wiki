package main

type stubStorage struct {
	page        wikiPage
	expectederr error
}

func (ss *stubStorage) storeFile(name string, content []byte) error {
	return nil
}

func (ss *stubStorage) getPublicPages() []string {
	return []string{}
}

func (ss *stubStorage) getPage(p *wikiPage) (*wikiPage, error) {
	return &ss.page, ss.expectederr
}

func (ss *stubStorage) searchPages(root, query string) []string {
	return []string{}
}