package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/darren-west/app/utils/httputil"
	"github.com/darren-west/app/utils/jwt"
	"github.com/go-resty/resty"
	"github.com/hashicorp/errwrap"
)

// WithBaseAddress is the base address for the client to use.
// An example is http://localhost/api
func WithBaseAddress(base string) Option {
	return func(s *Service) {
		if len(base) > 0 && base[len(base)-1] == '/' {
			s.base = base[:len(base)-1]
			return
		}
		s.base = base
	}
}

// WithRoundTripper sets the http round tripper to use in the client.
func WithRoundTripper(roundTripper http.RoundTripper) Option {
	return func(s *Service) {
		s.httpClient.SetTransport(roundTripper)
	}
}

// WithRetryCount sets the number of times to retry.
func WithRetryCount(retries int) Option {
	return func(s *Service) {
		s.httpClient.SetRetryCount(retries)
	}
}

// Option is a function for setting options on the Service.
type Option func(*Service)

// New returns a new client to the auth service.
func New(opts ...Option) Service {
	s := Service{
		httpClient: resty.New(),
		base:       "http://localhost/api/auth",
	}
	for _, opt := range opts {
		opt(&s)
	}
	s.httpClient.AddRetryCondition(resty.RetryConditionFunc(func(resp *resty.Response) (bool, error) {
		return resp.StatusCode() < 200 || resp.StatusCode() > 399, nil
	}))
	return s
}

// Service is a client to the auth service.
type Service struct {
	httpClient *resty.Client
	base       string
}

// ExchangeToken returns a signed jwt token for the user given. It is signed in the auth service.
func (s Service) ExchangeToken(ctx context.Context, user jwt.User) (token string, err error) {
	if token, err = s.exchangeToken(ctx, user); err != nil {
		err = errwrap.Wrapf("exchange token failed: {{err}}", err)
		return
	}
	return
}

func (s Service) exchangeToken(ctx context.Context, user jwt.User) (string, error) {
	resp, err := s.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(&user).
		SetContext(ctx).
		Post(s.pathf("/%s", "token"))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", httputil.NewError(resp.StatusCode()).WithMessage(string(resp.Body()))
	}
	return string(resp.Body()), nil
}

func (s Service) pathf(format string, args ...interface{}) string {
	return s.base + fmt.Sprintf(format, args...)
}
