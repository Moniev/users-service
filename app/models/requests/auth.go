package requests

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
	return nil
}

type Register struct {
	Name     string `json:"name"`
	Mail     string `json:"mail" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Device
}

func (r *Register) Valid() error {
	return nil
}

type Login struct {
	Mail     string
	Password string
	Device
}

func (r Login) Valid() error {
	return nil
}

type Code struct {
	Code string
	Device
}

func (r Code) Valid() error {
	return nil
}

type Mail struct {
	Mail string
	Device
}

func (r *Mail) Valid() error {
	return nil
}

type ConfirmPasswordReset struct {
	Code     string
	Password string
	Device
}

func (r *ConfirmPasswordReset) Valid() error {
	return nil
}
