package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/darren-west/app/utils/jwt"
)

func WithTargetAddress(target string) Option {
	return func(s *Service) {
		s.TargetAddress = target
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(s *Service) {
		s.httpClient = httpClient
	}
}

type Option func(*Service)

func New(opts ...Option) Service {
	s := Service{
		httpClient:    http.DefaultClient,
		TargetAddress: "http://localhost/api/auth",
	}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

type Service struct {
	httpClient    *http.Client
	TargetAddress string
}

func (s Service) ExchangeToken(ctx context.Context, user jwt.User) (string, error) {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&user); err != nil {
		return "", err
	}
	req, err := s.newRequest(http.MethodPost, "token", buf)
	if err != nil {
		return "", err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s Service) newRequest(method string, path string, body io.Reader) (req *http.Request, err error) {
	return http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.TargetAddress, path), body)
}
