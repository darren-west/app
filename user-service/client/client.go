package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/darren-west/app/user-service/models"
	"github.com/darren-west/app/utils/httputil"
	"github.com/go-resty/resty"
	"github.com/hashicorp/errwrap"
)

// WithBaseAddress sets the base address to communicate with.
// for example http://localhost/api
func WithBaseAddress(base string) Option {
	return func(s *Service) {
		if len(base) > 0 && base[len(base)-1] == '/' {
			s.base = base[:len(base)-1]
			return
		}
		s.base = base
	}
}

// WithRoundTripper sets the http round trip interface to use.
func WithRoundTripper(roundTripper http.RoundTripper) Option {
	return func(s *Service) {
		s.httpClient.SetTransport(roundTripper)
	}
}

// WithRetryCount sets the number of retries before failing.
func WithRetryCount(retries int) Option {
	return func(s *Service) {
		s.httpClient.SetRetryCount(retries)
	}
}

// Option is a function for setting the options in the service.
type Option func(*Service)

// New returns a new service client for the user service.
func New(ops ...Option) Service {
	s := Service{
		base:       "http://localhost",
		httpClient: resty.New(),
	}
	for _, op := range ops {
		op(&s)
	}
	s.httpClient.AddRetryCondition(resty.RetryConditionFunc(func(resp *resty.Response) (bool, error) {
		return resp.StatusCode() < 200 || resp.StatusCode() > 399, nil
	}))
	return s
}

// Service is a client to the user service.
type Service struct {
	httpClient *resty.Client
	base       string
}

func (s Service) pathf(format string, args ...interface{}) string {
	return s.base + fmt.Sprintf(format, args...)
}

// CreateUser creates a user.
func (s Service) CreateUser(ctx context.Context, user models.UserInfo) (err error) {
	err = func(ctx context.Context, user models.UserInfo) (err error) {
		resp, err := s.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetBody(&user).
			SetContext(ctx).
			Post(s.pathf("/%s", "users"))
		if err != nil {
			return
		}
		if httpErr := handleError(http.StatusCreated, resp); httpErr != nil {
			return httpErr
		}
		return
	}(ctx, user)
	if err != nil {
		return errwrap.Wrapf("create user failed: {{err}}", err)
	}
	return
}

// GetUser return the user with the id given.
func (s Service) GetUser(ctx context.Context, id string) (user models.UserInfo, err error) {
	user, err = func(ctx context.Context, id string) (user models.UserInfo, err error) {
		resp, err := s.httpClient.R().
			SetResult(&user).
			SetContext(ctx).
			Get(s.pathf("/%s/%s", "users", id))
		if httpErr := handleError(http.StatusOK, resp); httpErr != nil {
			return user, httpErr
		}
		return
	}(ctx, id)
	if err != nil {
		err = errwrap.Wrapf("get user failed: {{err}}", err)
		return
	}
	return
}

// IsNotFoundError returns true if the error is a not found error. i.e. the user is not found.
func IsNotFoundError(err error) bool {
	if e, ok := err.(httputil.Error); ok {
		if e.StatusCode() == http.StatusNotFound {
			return true
		}
	}
	return false
}

// ListUsers returns all the users.
// TODO: implement filtering.
func (s Service) ListUsers(ctx context.Context) (users []models.UserInfo, err error) {
	users, err = func(ctx context.Context) (users []models.UserInfo, err error) {
		resp, err := s.httpClient.R().
			SetResult(&users).
			SetContext(ctx).
			Get(s.pathf("/%s", "users"))
		if httpErr := handleError(http.StatusOK, resp); httpErr != nil {
			return nil, httpErr
		}
		return
	}(ctx)
	if err != nil {
		err = errwrap.Wrapf("list users failed: {{err}}", err)
		return
	}
	return
}

// UpdateUser updates the user. The ID in the user is used to update the user in the service.
func (s Service) UpdateUser(ctx context.Context, user models.UserInfo) (err error) {
	err = func(ctx context.Context, user models.UserInfo) (err error) {
		resp, err := s.httpClient.R().
			SetHeader("Content-Type", "application/json").
			SetBody(&user).
			SetContext(ctx).
			Put(s.pathf("/%s/%s", "users", user.ID))
		if httpErr := handleError(http.StatusOK, resp); httpErr != nil {
			return httpErr
		}
		return
	}(ctx, user)
	if err != nil {
		return errwrap.Wrapf("update user failed: {{err}}", err)
	}
	return
}

// DeleteUser removes the user in the service.
func (s Service) DeleteUser(ctx context.Context, id string) (err error) {
	err = func(ctx context.Context, id string) (err error) {
		resp, err := s.httpClient.R().SetContext(ctx).Delete(s.pathf("/%s/%s", "users", id))
		if httpErr := handleError(http.StatusOK, resp); httpErr != nil {
			return httpErr
		}
		return
	}(ctx, id)
	if err != nil {
		return errwrap.Wrapf("delete user failed: {{err}}", err)
	}
	return
}

func handleError(expected int, resp *resty.Response) httputil.Error {
	if resp.StatusCode() != expected {
		return httputil.NewError(resp.StatusCode()).WithMessage("%s, code %d", string(resp.Body()), resp.StatusCode())
	}
	return nil
}
