package requests

type Empty struct {
	Device
}

func (r *Empty) Valid() error {
	return translateError(validate.Struct(r))
}

type Details struct {
	FirstName string `json:"first_name" binding:"required" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" binding:"required" validate:"required,min=1,max=50"`
	Device
}

func (r *Details) Valid() error {
	return translateError(validate.Struct(r))
}

type User struct {
	Phone string `json:"phone" binding:"required" validate:"required,custom_phone"`
	Mail  string `json:"mail" binding:"required,email" validate:"required,custom_email"`
	Device
}

func (r *User) Valid() error {
	return translateError(validate.Struct(r))
}

type UserPublic struct {
	ID int `json:"id" binding:"required" validate:"required,gt=0"`
	Device
}

func (r *UserPublic) Valid() error {
	return translateError(validate.Struct(r))
}

type Settings struct {
	TwoFactor                    bool  `json:"two_factor"`
	NightMode                    bool  `json:"night_mode"`
	SecondFactorTargetID         int   `json:"second_factor_target_id" binding:"required" validate:"required,gt=0"`
	NotificationsTargetDeviceIDs []int `json:"notifications_target_device_ids" validate:"dive,gt=0"`
	Device
}

func (r *Settings) Valid() error {
	return translateError(validate.Struct(r))
}

type Password struct {
	Password string `json:"password" binding:"required,min=8" validate:"required,custom_password"`
	Device
}

func (r *Password) Valid() error {
	return translateError(validate.Struct(r))
}

type EntrepreneurDetails struct {
	BusinessName             string   `json:"business_name" binding:"required" validate:"required,min=2,max=100"`
	NIP                      string   `json:"nip" binding:"omitempty" validate:"custom_nip"`
	KRS                      string   `json:"krs" binding:"omitempty" validate:"custom_krs"`
	Description              string   `json:"description" binding:"omitempty" validate:"max=500"`
	Income                   float64  `json:"income" binding:"omitempty" validate:"gte=0"`
	Costs                    float64  `json:"costs" binding:"omitempty" validate:"gte=0"`
	FundingCapital           float64  `json:"funding_capital" binding:"omitempty" validate:"gte=0"`
	Industry                 string   `json:"industry" binding:"omitempty" validate:"max=100"`
	ManagementCouncilMembers []string `json:"management_council_members" binding:"omitempty" validate:"dive,min=1,max=100"`
	DecisionMakers           []string `json:"decision_makers" binding:"omitempty" validate:"dive,min=1,max=100"`
	BusinessPhoneNumber      string   `json:"business_phone_number" binding:"omitempty" validate:"custom_phone"`
	BusinessMail             string   `json:"business_mail" binding:"omitempty,email" validate:"custom_email"`
	WebsiteAddress           string   `json:"website_address" binding:"omitempty" validate:"custom_url"`
	Device
}

func (r *EntrepreneurDetails) Valid() error {
	return translateError(validate.Struct(r))
}

type Location struct {
	Country         string `json:"country" binding:"required" validate:"required,min=2,max=100"`
	Province        string `json:"province" binding:"required" validate:"required,max=100"`
	City            string `json:"city" binding:"required" validate:"required,min=2,max=100"`
	PostalCode      string `json:"postal_code" binding:"required" validate:"required,min=5,max=10"`
	Street          string `json:"street" binding:"omitempty" validate:"max=100"`
	BuildingNumber  string `json:"building_number" binding:"required" validate:"required,min=1,max=20"`
	ApartmentNumber string `json:"apartment_number" binding:"omitempty" validate:"max=20"`
	Device
}

func (r *Location) Valid() error {
	return translateError(validate.Struct(r))
}

type Index struct {
	Page     int `json:"page" binding:"required" validate:"required,gt=0"`
	PageSize int `json:"page_size" binding:"required" validate:"required,gt=0,lte=100"`
	Device
}

func (r *Index) Valid() error {
	return translateError(validate.Struct(r))
}

type Role struct {
	UserID int `json:"user_id" binding:"required" validate:"required,gt=0"`
	RoleID int `json:"role_id" binding:"required" validate:"required,gt=0"`
	Device
}

func (r *Role) Valid() error {
	return translateError(validate.Struct(r))
}
