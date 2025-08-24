package requests

import (
	"errors"
	"regexp"
	"sync"
	"users-service/app/utils"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func init() {
	once.Do(func() {
		validate = validator.New()
		validate.RegisterValidation("custom_email", func(fl validator.FieldLevel) bool {
			email := fl.Field().String()
			if email == "" {
				return true
			}
			return utils.CheckEmailFormat(email)
		})

		validate.RegisterValidation("custom_password", func(fl validator.FieldLevel) bool {
			return utils.CheckPasswordFormat(fl.Field().String())
		})

		validate.RegisterValidation("custom_token", func(fl validator.FieldLevel) bool {
			token := fl.Field().String()
			if token == "" {
				return true
			}
			return utils.CheckTokenFormat(token)
		})

		validate.RegisterValidation("custom_phone", func(fl validator.FieldLevel) bool {
			phone := fl.Field().String()
			if phone == "" {
				return true
			}
			return utils.CheckPhoneFormat(phone)
		})

		validate.RegisterValidation("custom_nip", func(fl validator.FieldLevel) bool {
			nip := fl.Field().String()
			if nip == "" {
				return true
			}

			matched, _ := regexp.MatchString(`^\d{10}$`, nip)
			return matched
		})

		validate.RegisterValidation("custom_krs", func(fl validator.FieldLevel) bool {
			krs := fl.Field().String()
			if krs == "" {
				return true
			}

			matched, _ := regexp.MatchString(`^\d{10,14}$`, krs)
			return matched
		})

		validate.RegisterValidation("custom_url", func(fl validator.FieldLevel) bool {
			url := fl.Field().String()
			if url == "" {
				return true
			}

			matched, _ := regexp.MatchString(`^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, url)
			return matched
		})
	})
}

type Device struct {
	DeviceToken    string `json:"device_token" binding:"required" validate:"required"`
	IPAddress      string `json:"ip_address" binding:"required" validate:"required,ip"`
	UserAgent      string `json:"user_agent" binding:"required" validate:"required"`
	OSName         string `json:"os_name" binding:"required" validate:"required"`
	OSVersion      string `json:"os_version" binding:"required" validate:"required"`
	BrowserName    string `json:"browser_name" binding:"required" validate:"required"`
	BrowserVersion string `json:"browser_version" binding:"required" validate:"required"`
}

type DeviceInterface interface {
	Valid() error
}

func (r Device) Valid() error {
	if err := validate.Struct(r); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				switch e.Field() {
				case "DeviceToken":
					return errors.New("no device token provided")
				case "IPAddress":
					return errors.New("no ip address provided")
				case "UserAgent":
					return errors.New("no user agent provided")
				case "OSName":
					return errors.New("no os name provided")
				case "OSVersion":
					return errors.New("no os version provided")
				case "BrowserName":
					return errors.New("no browser name provided")
				case "BrowserVersion":
					return errors.New("no browser version provided")
				}
			}
		}
		return err
	}
	return nil
}

type Register struct {
	Name     string `json:"name" binding:"required,min=3,max=50" validate:"required,min=3,max=50"`
	Mail     string `json:"mail" binding:"required,email" validate:"required,custom_email"`
	Password string `json:"password" binding:"required,min=8" validate:"required,custom_password"`
	Device
}

func (r *Register) Valid() error {
	return validate.Struct(r)
}

type Login struct {
	Mail     string `json:"mail" binding:"required,email" validate:"required,custom_email"`
	Password string `json:"password" binding:"required,min=8" validate:"required,custom_password"`
	Device
}

func (r *Login) Valid() error {
	return validate.Struct(r)
}

type Code struct {
	Code string `json:"code" binding:"required" validate:"required,custom_token"`
	Device
}

func (r *Code) Valid() error {
	return validate.Struct(r)
}

type Mail struct {
	Mail string `json:"mail" binding:"required,email" validate:"required,custom_email"`
	Device
}

func (r *Mail) Valid() error {
	return validate.Struct(r)
}

type ConfirmPasswordReset struct {
	Code     string `json:"code" binding:"required" validate:"required,custom_token"`
	Password string `json:"password" binding:"required,min=8" validate:"required,custom_password"`
	Device
}

func (r *ConfirmPasswordReset) Valid() error {
	return validate.Struct(r)
}
