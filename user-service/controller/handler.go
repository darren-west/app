package controller

// TODO: do some error handling to test for duplicate entires and other common errors. Print nicer messages

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/darren-west/app/utils/httputil"

	"github.com/darren-west/app/user-service/models"
	"github.com/darren-west/app/user-service/repository"
	"github.com/julienschmidt/httprouter"
)

func handleError(err error) httputil.Error {
	if repository.IsErrUserNotFound(err) {
		return httputil.NewError(http.StatusNotFound).WithError(err)
	}
	return httputil.NewError(http.StatusInternalServerError).WithError(err)
}

func NewHandler(us UserRepository, r *httprouter.Router) http.Handler {
	h := Handler{UserRepository: us, UserValidator: models.UserValidator{}}
	r.GET("/users/:id", UseErrorHandle(h.GetUser))
	r.GET("/users", UseErrorHandle(h.ListUsers))
	r.DELETE("/users/:id", UseErrorHandle(h.DeleteUser))
	r.PUT("/users/:id", UseErrorHandle(h.UpdateUser))
	r.POST("/users", UseErrorHandle(h.CreateUser))
	return ensureContentType(r)
}

func ensureContentType(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method == http.MethodPost || r.Method == http.MethodPut) && r.Header.Get("Content-Type") != "application/json" {
			httputil.NewError(http.StatusUnsupportedMediaType).WithMessage("unsupported content type").Write(w)
			return
		}
		h.ServeHTTP(w, r)
	}
}

type Handler struct {
	UserRepository
	models.UserValidator
}

func (h Handler) GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) httputil.Error {
	user, err := h.UserRepository.FindUser(repository.NewMatcher().WithID(ps.ByName("id")))
	if err != nil {
		return handleError(err)
	}
	if err = encodeJSON(w, &user, isPretty(r)); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	return nil
}

func (h Handler) ListUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) httputil.Error {
	users, err := h.UserRepository.ListUsers(repository.EmptyMatcher)
	if err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	if err = encodeJSON(w, &users, isPretty(r)); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	return nil
}

func isPretty(r *http.Request) (pretty bool) {
	_, pretty = r.URL.Query()["pretty"]
	return
}

func encodeJSON(w io.Writer, i interface{}, pretty bool) error {
	e := json.NewEncoder(w)
	if pretty {
		e.SetIndent("", "\t")
	}
	return e.Encode(i)
}

func (h Handler) DeleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) httputil.Error {
	if err := h.UserRepository.RemoveUser(repository.NewMatcher().WithID(ps.ByName("id"))); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	return nil
}

func (h Handler) UpdateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) httputil.Error {
	user := models.UserInfo{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	if user.ID != ps.ByName("id") {
		return httputil.NewError(http.StatusBadRequest).WithMessage("query param id (%s) does not match request body id (%s)", user.ID, ps.ByName("id"))
	}
	if err := h.UserValidator.IsValid(user); err != nil {
		return httputil.NewError(http.StatusBadRequest).WithError(err)
	}
	if err := h.UserRepository.UpdateUser(user); err != nil {
		return handleError(err)
	}
	return nil
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) httputil.Error {
	user := models.UserInfo{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	if err := h.UserValidator.IsValid(user); err != nil {
		return httputil.NewError(http.StatusBadRequest).WithError(err)
	}
	if err := h.UserRepository.CreateUser(user); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	w.WriteHeader(http.StatusCreated)
	return nil
}

type ErrorHandle func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err httputil.Error)

func UseErrorHandle(f ErrorHandle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := f(w, r, ps); err != nil {
			err.Write(w)
			return
		}

	}
}

//go:generate mockgen -destination ./mocks/mock_service.go -package mocks github.com/darren-west/app/user-service/controller UserRepository

type UserRepository interface {
	FindUser(repository.Matcher) (models.UserInfo, error)
	ListUsers(repository.Matcher) ([]models.UserInfo, error)
	RemoveUser(repository.Matcher) error
	UpdateUser(models.UserInfo) error
	CreateUser(models.UserInfo) error
}
