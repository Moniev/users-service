package responses

import (
	"errors"
	"users-service/app/models/ent"
)

type User struct {
	User  *ent.User `json:"user"`
	Token string    `json:"token,omitempty"`
}

type UserInterface interface {
	Valid() error
}

func (r *User) Valid() error {
	if r.User == nil {
		return errors.New("no user provided")
	}

	return nil
}

func FromUser(user *ent.User) *User {
	return &User{}
}
