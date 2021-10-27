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

// OpenFile opens a file for writing, and if the file exists and is bigger than maxSize it'll truncate it first
func (fm *FileManager) OpenFile(filePath string, maxSize int64) (*os.File, error) {
	var logfileMode int

	if info, err := os.Stat(filePath); os.IsNotExist(err) {
		logfileMode = os.O_APPEND
	} else if err != nil {
		log.WithError(err).WithField("file", filePath).Fatal("cannot stat logfile")
	} else if info.Size() >= maxSize {
		logfileMode = os.O_TRUNC
	} else {
		logfileMode = os.O_APPEND
	}

	return os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|logfileMode, 0644)
}
