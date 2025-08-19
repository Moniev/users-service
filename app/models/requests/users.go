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
	if r.FirstName == "" || r.LastName == "" {
		return errors.New("")
	}

	return r.Device.Valid()
}

type User struct {
	Phone string `json:"phone"`
	Mail  string `json:"mail"`
	Device
}

type UserPublic struct {
	ID int `json:"id"`
	Device
}

type UserInterface interface {
	Valid() error
}

func (r *User) Valid() error {
	if valid := utils.CheckEmailFormat(r.Mail); !valid && r.Mail != "" {
		return errors.New("provided not valid email format")
	}

	if valid := utils.CheckPhoneFormat(r.Phone); !valid && r.Phone != "" {
		return errors.New("provided not valid phone format")
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
	if r.SecondFactorTargetID <= 0 {
		return errors.New("wrong second factor device ID provided")
	}

	for _, id := range r.NotificationsTargetDeviceIDs {
		if id <= 0 {
			return errors.New("wrong notification device ID provided")
		}
	}

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

type EntrepreneurDetails struct {
	BusinessName             string   `json:"business_name"`
	NIP                      string   `json:"nip"`
	KRS                      string   `json:"krs"`
	Description              string   `json:"description"`
	Income                   float64  `json:"income"`
	Costs                    float64  `json:"costs"`
	FundingCapital           float64  `json:"funding_capital"`
	Industry                 string   `json:"industry"`
	ManagementCouncilMembers []string `json:"management_council_members"`
	DecisionMakers           []string `json:"decision_makers"`
	BusinessPhoneNumber      string   `json:"business_phone_number"`
	BusinessMail             string   `json:"business_mail"`
	WebsiteAddress           string   `json:"website_address"`
	Device
}

func (r *EntrepreneurDetails) Valid() error {
	if r.BusinessPhoneNumber != "" && !utils.CheckPhoneFormat(r.BusinessPhoneNumber) {
		return errors.New("provided wrong format of phone number")
	}

	return r.Device.Valid()
}

type Location struct {
	Country         string `json:"country"`
	Province        string `json:"province"`
	City            string `json:"city"`
	PostalCode      string `json:"postal_code"`
	Street          string `json:"street"`
	BuildingNumber  string `json:"building_number"`
	ApartmentNumber string `json:"apartment_number"`
	Device
}

func (r *Location) Valid() error {
	return r.Device.Valid()
}

type Index struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Device
}

func (r *Index) Valid() error {
	if r.PageSize > 0 && r.Page > 0 {
		return errors.New("choosed wrong page size nad pages number")
	}

	return r.Device.Valid()
}

type Role struct {
	UserID int
	RoleID int
	Device
}

func (r *Role) Valid() error {
	if r.UserID <= 0 {
		return errors.New("choosed wrong user ID")
	}

	if r.RoleID <= 0 {
		return errors.New("choosed wrong role ID")
	}

	return r.Device.Valid()
}
