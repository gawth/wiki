package main

import "log"

type cachedStorage struct {
	fileStorage
	wikiDir        string
	tagDir         string
	cachedTagIndex TagIndex
	cachedRawFiles TagIndex
}

func newCachedStorage(fs fileStorage, wd, td string) cachedStorage {
	ti := fs.IndexTags(td)
	rf := fs.IndexRawFiles(wd, "PDF", ti)

	return cachedStorage{fs, wd, td, ti, rf}
}

func (cs *cachedStorage) IndexTags(path string) TagIndex {
	if cs.cachedTagIndex == nil {
		log.Println("[cache] cache rebuild")
		cs.cachedTagIndex = cs.fileStorage.IndexTags(cs.tagDir)
		cs.cachedRawFiles = cs.fileStorage.IndexRawFiles(cs.wikiDir, "PDF", cs.cachedTagIndex)
	}
	return cs.cachedTagIndex
}

func (cs *cachedStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	return cs.cachedRawFiles
}

func (cs *cachedStorage) clearCache() {
	cs.cachedRawFiles = nil
	cs.cachedTagIndex = nil
}
func (cs *cachedStorage) storeFile(name string, content []byte) error {
	cs.clearCache()
	return cs.fileStorage.storeFile(name, content)
}
func (cs *cachedStorage) deleteFile(name string) error {
	cs.clearCache()
	return cs.fileStorage.deleteFile(name)
}
func (cs *cachedStorage) moveFile(from, to string) error {
	cs.clearCache()
	return cs.fileStorage.moveFile(from, to)
}
