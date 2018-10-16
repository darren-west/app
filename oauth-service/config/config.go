package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
)

//go:generate mockgen -destination ./mocks/mock_reader.go -package mocks github.com/darren-west/app/oauth-service/config FileReader

// DefaultFileReader wraps ioutil.ReadFile call to read a file into a byte array.
type DefaultFileReader struct{}

// Read reads the file at the path given.
func (DefaultFileReader) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// FileReader reads a files content into a byte array.
type FileReader interface {
	Read(string) ([]byte, error)
}

// Reader is used to read configuration in from a file. Instantiate using the NewReader function.
type Reader struct {
	fileReader FileReader
}

// Read reads the config in the file and returns the oauth2 config.
func (r Reader) Read(path string) (config *oauth2.Config, err error) {
	data, err := r.fileReader.Read(path)
	if err != nil {
		err = fmt.Errorf("failed to read file: %s", err)
		return
	}
	config = new(oauth2.Config)
	if err = json.Unmarshal(data, config); err != nil {
		return
	}
	return
}

// NewReader creates a new reader using the injected file reader to read the contents of the file.
func NewReader(fileReader FileReader) (Reader, error) {
	if fileReader == nil {
		return Reader{}, errors.New("file reader is nil")
	}
	return Reader{fileReader: fileReader}, nil
}
