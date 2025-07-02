package requests

import (
	"errors"
	"users-service/app/utils"
)

type Device struct {
	DeviceToken    string `json:"device_token" binding:"required"`
	IPAddress      string `json:"ip_address"`
	UserAgent      string `json:"user_agent"`
	OSName         string `json:"os_name"`
	OSVersion      string `json:"os_version"`
	BrowserName    string `json:"browser_name"`
	BrowserVersion string `json:"browser_version"`
}

type DeviceInterface interface {
	Valid() error
}

func (r Device) Valid() error {
	if r.IPAddress == "" {
		return errors.New("no ip address provided")
	}

	if r.OSName == "" {
		return errors.New("no os name provided")
	}

	if r.OSVersion == "" {
		return errors.New("no os version provided")
	}

	if r.BrowserName == "" {
		return errors.New("no browser name provided")
	}

	if r.OSVersion == "" {
		return errors.New("no browser version provided")
	}

	return nil
}

type Register struct {
	Name     string `json:"name"`
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Device
}

func (r *Register) Valid() error {
	if !utils.CheckEmailFormat(r.Mail) {
		return errors.New("invalid email format")
	}

	if !utils.CheckPasswordFormat(r.Password) {
		return errors.New("password does not meet complexity requirements")
	}

	return r.Device.Valid()
}

type Login struct {
	Mail     string `json:"mail"`
	Password string `json:"password"`
	Device
}

func (r Login) Valid() error {
	if !utils.CheckEmailFormat(r.Mail) {
		return errors.New("invalid email format provided")
	}

	if !utils.CheckPasswordFormat(r.Password) {
		return errors.New("password does not meet complexity requirements")
	}

	return r.Device.Valid()
}

type Code struct {
	Code string `json:"code"`
	Device
}

func (r Code) Valid() error {
	if !utils.CheckTokenFormat(r.Code) {
		return errors.New("invalid code format provided")
	}

	return r.Device.Valid()
}

type Mail struct {
	Mail string `json:"mail"`
	Device
}

func (r *Mail) Valid() error {
	if !utils.CheckEmailFormat(r.Mail) {
		return errors.New("wrong mail provided")
	}

	return r.Device.Valid()
}

type ConfirmPasswordReset struct {
	Code     string `json:"code"`
	Password string `json:"password"`
	Device
}

func (r *ConfirmPasswordReset) Valid() error {
	if !utils.CheckTokenFormat(r.Code) {
		return errors.New("wrong code provided")
	}

	if !utils.CheckPasswordFormat(r.Password) {
		return errors.New("wrong password provided")
	}

	err := r.Device.Valid()
	return err
}
