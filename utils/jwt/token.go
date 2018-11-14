package jwt

import (
	"errors"
	"fmt"

	"github.com/darren-west/app/utils/fileutil"
	jwt "github.com/dgrijalva/jwt-go"
)

// WriterBuilder is a type for setting options in the Writer.
var WriterBuilder = writerBuilder{}

type writerBuilder struct{}

// WithPrivateKeyPath sets the path of the private key to sign the token with,.
func (writerBuilder) WithPrivateKeyPath(path string) WriterOption {
	return func(w *Writer) {
		w.privateKeyPath = path
	}
}

// WithFileReader sets the file reader to implementation to read the private key with.
func (writerBuilder) WithFileReader(reader FileReader) WriterOption {
	return func(w *Writer) {
		if reader == nil {
			panic(errors.New("file reader is nil"))
		}
		w.fileReader = reader
	}
}

// WriterOption is used to set options on the writer.
type WriterOption func(*Writer)

// NewWriter creates a new writer with default options overriden by the options passed as arguments.
func NewWriter(opts ...WriterOption) Writer {
	w := Writer{
		privateKeyPath: "app.rsa",
		fileReader:     fileutil.FileReader{},
	}
	for _, op := range opts {
		op(&w)
	}
	return w
}

//go:generate mockgen -destination ./mocks/mock_reader.go -package mocks github.com/darren-west/app/auth-service/jwt FileReader

// FileReader is an interface for reading from a file.
type FileReader interface {
	Read(string) ([]byte, error)
}

// NewToken create a new token type from the string passed in.
func NewToken(tokenString string) Token {
	return Token(tokenString)
}

// Token is a type to represent a JWT token.
type Token string

// String returns the string representation of the JWT token.
func (t Token) String() string {
	return string(t)
}

// Writer is a type used to write
type Writer struct {
	privateKeyPath string
	fileReader     FileReader
}

// Write is a function for writing a signed JWT token with the claims passed in.
func (w Writer) Write(c *Claims) (Token, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	b, err := w.fileReader.Read(w.privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to write token: %s", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(b)
	if err != nil {
		return "", fmt.Errorf("failed to write token: %s", err)
	}
	token.Claims = c
	tokenString, err := token.SignedString(key)
	return Token(tokenString), err
}

// PrivateKeyPath is the path to the private key that the writer is using.
func (w Writer) PrivateKeyPath() string {
	return w.privateKeyPath
}

// FileReader is the file reader that is being used.
func (w Writer) FileReader() FileReader {
	return w.fileReader
}

// ReaderOption is a type for setting the options on a reader.
type ReaderOption func(*Reader)

// ReaderBuilder is used to create a Reader with non default options.
var ReaderBuilder = readerBuilder{}

type readerBuilder struct{}

// WithPublicKeyPath sets the path to the public key.
func (readerBuilder) WithPublicKeyPath(path string) ReaderOption {
	return func(r *Reader) {
		r.publicKeyPath = path
	}
}

// WithFileReader sets the file reader to use.
func (readerBuilder) WithFileReader(reader FileReader) ReaderOption {
	return func(r *Reader) {
		r.fileReader = reader
	}
}

// NewReader creates a reader. The default options are overriden by any passed in.
func NewReader(opts ...ReaderOption) Reader {
	r := Reader{
		fileReader:    fileutil.FileReader{},
		publicKeyPath: "app.rsa.pub",
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

// Reader reads JWT tokens and returns the claims.
type Reader struct {
	publicKeyPath string
	fileReader    FileReader
}

// Read reads a signed Token and returns the Claims.
func (r Reader) Read(t Token) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t.String(), new(Claims), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		data, err := r.fileReader.Read(r.publicKeyPath)
		if err != nil {
			return nil, err
		}
		pubKey, err := jwt.ParseRSAPublicKeyFromPEM(data)
		if err != nil {
			return nil, err
		}

		return pubKey, nil
	})
	return token.Claims.(*Claims), err
}

// PublicKeyPath returns the path to the public key.
func (r Reader) PublicKeyPath() string {
	return r.publicKeyPath
}

// FileReader returns the file reader used to read the public key.
func (r Reader) FileReader() FileReader {
	return r.fileReader
}

// Claims are the claims assigned to the JWT token.
type Claims struct {
	User      User  `json:"user,omitempty"`
	ExpiresAt int64 `json:"exp,omitempty"`
	IssuedAt  int64 `json:"iat,omitempty"`
}

func (*Claims) Valid() (err error) {
	return
}

type User struct {
	ID        string `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}
