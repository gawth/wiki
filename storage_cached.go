package main

import "log"

type cachedStorage struct {
	fileStorage
	wikiDir         string
	tagDir          string
	cachedTagIndex  TagIndex
	cachedRawFiles  TagIndex
	cachedWikiIndex []wikiNav
}

func newCachedStorage(fs fileStorage, wd, td string) cachedStorage {
	ti := fs.IndexTags(td)
	rf := fs.IndexRawFiles(wd, "PDF", ti)
	wi := fs.IndexWikiFiles("", wd)

	return cachedStorage{fs, wd, td, ti, rf, wi}
}

func (cs *cachedStorage) rebuildCache() {
	log.Println("[cache] wiki cache rebuild")
	cs.cachedTagIndex = cs.fileStorage.IndexTags(cs.tagDir)
	cs.cachedRawFiles = cs.fileStorage.IndexRawFiles(cs.wikiDir, "PDF", cs.cachedTagIndex)
	cs.cachedWikiIndex = cs.fileStorage.IndexWikiFiles("", cs.wikiDir)
}

func (cs *cachedStorage) IndexWikiFiles(base, path string) []wikiNav {
	return cs.cachedWikiIndex
}

func (cs *cachedStorage) IndexTags(path string) TagIndex {
	return cs.cachedTagIndex
}

func (cs *cachedStorage) IndexRawFiles(path, fileExtension string, existing TagIndex) TagIndex {
	return cs.cachedRawFiles
}

func (cs *cachedStorage) clearCache() error {
	// Original async implementation - rebuilds cache in background
	go cs.rebuildCache()
	return nil
}
func (cs *cachedStorage) storeFile(name string, content []byte) error {
	if err := cs.fileStorage.storeFile(name, content); err != nil {
		return err
	}
	return cs.clearCache()
}
func (cs *cachedStorage) deleteFile(name string) error {
	if err := cs.fileStorage.deleteFile(name); err != nil {
		return err
	}
	return cs.clearCache()
}
func (cs *cachedStorage) moveFile(from, to string) error {
	if err := cs.fileStorage.moveFile(from, to); err != nil {
		return err
	}
	return cs.clearCache()
}
