package requests

import (
	"errors"
	"users-service/app/utils"
)

type Empty struct {
	Device
}

type EmptyInterface interface {
	Valid() error
}

func (r *Empty) Valid() error {
	return r.Device.Valid()
}

type Details struct {
	FirstName string
	LastName  string
	Device
}

type DetailsInterface interface {
	Valid() error
}

func (r *Details) Valid() error {
	return r.Device.Valid()
}

type User struct {
	Phone string `json:"phone"`
	Mail  string `json:"mail"`
	Device
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

	return r.Device.Valid()
}

type Settings struct {
	TwoFactor                    bool  `json:"two_factor"`
	NightMode                    bool  `json:"night_mode"`
	SecondFactorTargetID         int   `json:"second_factor_target_id"`
	NotificationsTargetDeviceIDs []int `json:"notifications_target_device_ids"`
	Device
}

type SettingsInerface interface {
	Valid() error
}

func (r *Settings) Valid() error {
	return r.Device.Valid()
}

type Password struct {
	Password string `json:"password"`
	Device
}

type PasswordInterface interface {
	Valid() error
}

func (r *Password) Valid() error {
	if !utils.CheckPasswordFormat(r.Password) {
		return errors.New("wrong password provided")
	}

	return r.Device.Valid()
}
