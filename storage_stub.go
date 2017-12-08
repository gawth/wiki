package main

type stubStorage struct {
	page            wikiPage
	expectederr     error
	GetTagWikisFunc func(string) Tag
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

func (ss *stubStorage) checkForPDF(p *wikiPage) (*wikiPage, error) {
	return &ss.page, ss.expectederr
}
func (ss *stubStorage) IndexTags(path string) TagIndex {
	return nil
}
func (ss *stubStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	return nil
}
func (ss *stubStorage) GetTagWikis(tag string) Tag {
	return ss.GetTagWikisFunc(tag)
}
