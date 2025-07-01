package requests

import (
	"errors"
	"users-service/app/utils"
)

type Empty struct {
}

type EmptyInterface interface {
	Valid() error
}

func (r *Empty) Valid() error {
	return nil
}

type Details struct {
}

type DetailsInterface interface {
	Valid() error
}

func (r *Details) Valid() error {
	return nil
}

type User struct {
	Phone string
	Mail  string
}

type UserInterface interface {
	Valid() error
}

func (r *User) Valid() error {
	if valid := utils.CheckEmailFormat(r.Mail); !valid {
		return errors.New("prvided not valid email format")
	}

	if valid := utils.CheckPhoneFormat(r.Phone); !valid {
		return errors.New("prvided not valid phone format")
	}

	return nil
}

type Settings struct {
	TwoFactor                    bool
	NightMode                    bool
	SecondFactorTargetID         int
	NotificationsTargetDeviceIDs []int
}

type SettingsInerface interface {
	Valid() error
}

func (r *Settings) Valid() error {
	return nil
}

type Password struct {
	Password string
}

type PasswordInterface interface {
	Valid() error
}

func (r *Password) Valid() error {
	return nil
}
