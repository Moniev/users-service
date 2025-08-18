package responses

import (
	"time"
	"users-service/app/models/ent"
)

type UserPublic struct {
	ID          int       `json:"id"`
	Active      bool      `json:"active"`
	Verified    bool      `json:"verified"`
	Blacklisted bool      `json:"blacklisted"`
	Removed     bool      `json:"removed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func FromUserPublic(user *ent.User) *UserPublic {
	return &UserPublic{
		ID:          user.ID,
		Active:      user.Active,
		Verified:    user.Verified,
		Removed:     user.Removed,
		Blacklisted: user.Blacklisted,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

type Users struct {
	Users []UserPublic
}

func FromUsers(users []*ent.User) *Users {
	res := make([]UserPublic, 0)

	for _, user := range users {
		pub := UserPublic{
			ID:          user.ID,
			Active:      user.Active,
			Verified:    user.Verified,
			Removed:     user.Removed,
			Blacklisted: user.Blacklisted,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
		res = append(res, pub)
	}

	return &Users{Users: res}
}

type UserPrivate struct {
	Token string `json:"token,omitempty"`
	Mail  string `json:"mail"`
	Phone string `json:"phone,omitempty"`
	UserPublic
}

func FromUserPrivate(user *ent.User, token ...string) *UserPrivate {
	if user == nil {
		return nil
	}

	res := &UserPrivate{
		Mail:  user.Mail,
		Phone: user.Phone,
	}

	res.UserPublic = UserPublic{
		ID:          user.ID,
		Active:      user.Active,
		Verified:    user.Verified,
		Removed:     user.Removed,
		Blacklisted: user.Blacklisted,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	if len(token) > 0 {
		res.Token = token[0]
	}

	return res
}
