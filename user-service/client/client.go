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

func WithHTTPClient(c *http.Client) Option {
	return func(s *Service) {
		s.Client = c
	}
}

type Option func(*Service)

func New(ops ...Option) Service {
	s := Service{
		options: Options{
			TargetAddress: "http://localhost",
		},
		Client: http.DefaultClient,
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
	options Options
	*http.Client
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
	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}
	if err = HandleError(http.StatusCreated, resp); err != nil {
		return
	}
	return
}

func (s Service) ListUsers(ctx context.Context) (users []models.UserInfo, err error) {
	req, err := s.newRequest(ctx, http.MethodGet, "users", nil)
	if err != nil {
		return
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}
	if err = HandleError(http.StatusOK, resp); err != nil {
		return
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
	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}
	if err = HandleError(http.StatusOK, resp); err != nil {
		return
	}
	return
}

func (s Service) DeleteUser(ctx context.Context, id string) (err error) {
	req, err := s.newRequest(ctx, http.MethodDelete, fmt.Sprintf("users/%s", id), nil)
	if err != nil {
		return
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}
	if err = HandleError(http.StatusOK, resp); err != nil {
		return
	}
	return
}

func HandleError(expected int, resp *http.Response) (err error) {
	if resp.StatusCode != expected {
		if resp.Body != nil {
			data, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				message := string(data)
				return fmt.Errorf("status code %d, message %s", resp.StatusCode, message)
			}
		}
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return
}

func (s Service) newRequest(ctx context.Context, method string, path string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", s.options.TargetAddress, path), r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req.WithContext(ctx), nil
}
