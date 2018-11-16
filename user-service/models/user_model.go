package models

import (
	"fmt"
)

type UserInfo struct {
	ID        string `json:"ID"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Email     string `json:"Email"`
}

type UserValidator struct{}

func (UserValidator) IsValid(us UserInfo) (err error) {
	if us.ID == "" {
		err = fmt.Errorf("invalid user, missing id")
		return
	}
	if us.Email == "" {
		err = fmt.Errorf("invalid user, missing email")
		return
	}
	if us.FirstName == "" {
		err = fmt.Errorf("invalid user, missing first name")
		return
	}
	if us.LastName == "" {
		err = fmt.Errorf("invalid user, missing last name")
		return
	}
	return
}
