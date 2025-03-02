package main

type stubStorage struct {
	page                wikiPage
	expectederr         error
	GetTagWikisFunc     func(string) Tag
	getPageFunc         func(*wikiPage) (*wikiPage, error)
	storeFileFunc       func(string, []byte) error
	storeImageFunc      func(string, []byte, string) (string, error)
	storeResizedImageFunc func(string, []byte, string, int, int) (string, error)
	loggerFunc          func(string)
}

func (ss *stubStorage) logit(method string) {
	if ss.loggerFunc != nil {
		ss.loggerFunc(method)
	}
}

func (ss *stubStorage) storeFile(name string, content []byte) error {
	return ss.storeFileFunc(name, content)
}
func (ss *stubStorage) deleteFile(name string) error {
	ss.logit("deleteFile")
	return nil
}
func (ss *stubStorage) moveFile(from, to string) error {
	ss.logit("moveFile")
	return nil
}
func (ss *stubStorage) getWikiList(from string) []string {
	ss.logit("getWikiList")
	return []string{"test1", "test2"}
}

func (ss *stubStorage) getPublicPages() []string {
	return []string{}
}

func (ss *stubStorage) getPage(p *wikiPage) (*wikiPage, error) {
	return ss.getPageFunc(p)
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
func (ss *stubStorage) IndexWikiFiles(base, path string) []wikiNav {
	return nil
}

func (ss *stubStorage) storeImage(wikiTitle string, imageData []byte, extension string) (string, error) {
	if ss.storeImageFunc != nil {
		return ss.storeImageFunc(wikiTitle, imageData, extension)
	}
	ss.logit("storeImage")
	return "/wiki/raw/images/" + wikiTitle + "/test.png", nil
}

func (ss *stubStorage) storeResizedImage(wikiTitle string, imageData []byte, extension string, width, height int) (string, error) {
	if ss.storeResizedImageFunc != nil {
		return ss.storeResizedImageFunc(wikiTitle, imageData, extension, width, height)
	}
	ss.logit("storeResizedImage")
	return "/wiki/raw/images/" + wikiTitle + "/test_resized.png", nil
}
