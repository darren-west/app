package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/darren-west/app/user-service/models"
)

func WithTargetAddress(target string) Option {
	return func(s *Service) {
		s.options.TargetAddress = target
	}
}

func WithHTTPClient(c *http.Client) Option {
	return func(s *Service) {
		s.httpClient = c
	}
}

type Option func(*Service)

func New(ops ...Option) Service {
	s := Service{
		options: Options{
			TargetAddress: "http://localhost",
		},
		httpClient: http.DefaultClient,
	}
	for _, op := range ops {
		op(&s)
	}
	return s
}

type Options struct {
	TargetAddress string
}

type Service struct {
	options    Options
	httpClient *http.Client
}

func (s Service) CreateUser(ctx context.Context, user models.UserInfo) (err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(&user); err != nil {
		return
	}
	req, err := s.newRequest(ctx, http.MethodPost, "users", buf)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	if httpErr := HandleError(http.StatusCreated, resp); httpErr != nil {
		return httpErr
	}
	return
}

func (s Service) GetUser(ctx context.Context, id string) (user models.UserInfo, err error) {
	req, err := s.newRequest(ctx, http.MethodGet, fmt.Sprintf("users/%s", id), nil)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	if httpErr := HandleError(http.StatusOK, resp); httpErr != nil {
		return user, httpErr
	}
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return
	}
	return
}

func IsNotFoundError(err error) bool {
	if e, ok := err.(*Error); ok {
		if e.StatusCode == http.StatusNotFound {
			return true
		}
	}
	return false
}

func (s Service) ListUsers(ctx context.Context) (users []models.UserInfo, err error) {
	req, err := s.newRequest(ctx, http.MethodGet, "users", nil)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	if httpErr := HandleError(http.StatusOK, resp); httpErr != nil {
		return nil, httpErr
	}
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return
	}
	return
}

func (s Service) UpdateUser(ctx context.Context, user models.UserInfo) (err error) {
	buf := new(bytes.Buffer)
	if err = json.NewEncoder(buf).Encode(&user); err != nil {
		return
	}
	req, err := s.newRequest(ctx, http.MethodPut, fmt.Sprintf("users/%s", user.ID), buf)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	if httpErr := HandleError(http.StatusOK, resp); httpErr != nil {
		return httpErr
	}
	return
}

func (s Service) DeleteUser(ctx context.Context, id string) (err error) {
	req, err := s.newRequest(ctx, http.MethodDelete, fmt.Sprintf("users/%s", id), nil)
	if err != nil {
		return
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	if httpErr := HandleError(http.StatusOK, resp); httpErr != nil {
		return httpErr
	}
	return
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("status code %d", e.StatusCode)
	}
	return fmt.Sprintf("status code %d, message %s", e.StatusCode, e.Message)
}

func HandleError(expected int, resp *http.Response) *Error {
	if resp.StatusCode != expected {
		httpErr := &Error{
			StatusCode: resp.StatusCode,
		}
		if resp.Body != nil {
			data, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				httpErr.Message = string(data)
			}
		}
		return httpErr
	}
	return nil
}

func (s Service) newRequest(ctx context.Context, method string, path string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", s.options.TargetAddress, path), r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(ctx), nil
}
