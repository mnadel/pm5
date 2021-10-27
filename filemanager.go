package main

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type FileManager struct {
}

// Mkdirs will ensure all directories leading up to `file` exist
func (fm *FileManager) Mkdirs(file string) error {
	dir := filepath.Dir(file)
	log.WithField("dir", dir).Info("ensuring directory")
	return os.MkdirAll(dir, 0755)
}
