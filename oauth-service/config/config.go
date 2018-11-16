package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/darren-west/app/utils/session"
	"github.com/darren-west/app/utils/validator"

	"golang.org/x/oauth2"
)

//go:generate mockgen -destination ./mocks/mock_reader.go -package mocks github.com/darren-west/app/oauth-service/config FileReader

// FileReader reads a files content into a byte array.
type FileReader interface {
	Read(string) ([]byte, error)
}

// Reader is used to read configuration in from a file. Instantiate using the NewReader function.
type Reader struct {
	fileReader FileReader
}

// Read reads the config in the file and returns the oauth2 config.
func (r Reader) Read(path string) (config Options, err error) {
	//config = defaultConfig
	data, err := r.fileReader.Read(path)
	if err != nil {
		err = fmt.Errorf("failed to read file: %s", err)
		return
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return
	}
	if err = validator.Default.IsValid(&config); err != nil {
		err = fmt.Errorf("configuration invalid: %s", err)
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

// Options is a struct containing the options for configuring the service.
type Options struct {
	BindAddress       string
	LoginRoutePath    string
	RedirectRoutePath string
	OAuth             *oauth2.Config
	MongoSession      session.Options
	APIEndpoint       string
	UserMapping       UserMapping
}

func (o Options) IsValid() (err error) {
	if o.BindAddress == "" {
		return errors.New("required field bind address missing")
	}
	if o.LoginRoutePath == "" {
		return errors.New("required field login route path missing")
	}
	if o.RedirectRoutePath == "" {
		return errors.New("required field redirect route path missing")
	}
	if o.OAuth == nil {
		return errors.New("required field oauth missing")
	}
	if o.APIEndpoint == "" {
		return errors.New("required field api endpoint missing")
	}
	return
}

// UserMapping are the options for configuring the marshalling of the returned user.
type UserMapping struct {
	ID           string
	FirstName    string
	LastName     string
	EmailAddress string
}

func (um UserMapping) IsValid() (err error) {
	if um.ID == "" {
		return errors.New("required field user mapping id missing")
	}
	if um.FirstName == "" {
		return errors.New("required field user mapping first name missing")
	}
	if um.LastName == "" {
		return errors.New("required field user mapping last name missing")
	}
	if um.EmailAddress == "" {
		return errors.New("required field user mapping email missing")
	}
	return
}
