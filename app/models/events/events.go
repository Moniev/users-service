package events

import (
	"time"
)

type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    int       `json:"user_id"`
}

type RegistrationEvent struct {
	BaseEvent
	Email          string    `json:"email"`
	ActivationCode string    `json:"activation_code"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type SecondFactorEvent struct {
	BaseEvent
	SecondFactorCode string    `json:"second_factor_code"`
	ExpiresAt        time.Time `json:"expires_at"`
	TargetDevice     string    `json:"target_device"`
}

type LoginEvent struct {
	BaseEvent
	LoginMethod   string   `json:"login_method"`
	OriginDevice  string   `json:"origin_device,omitempty"`
	TargetDevices []string `json:"target_devices,omitempty"`
}

type LogoutEvent struct {
	BaseEvent
	LoginMethod   string   `json:"login_method"`
	OriginDevice  string   `json:"origin_device,omitempty"`
	TargetDevices []string `json:"target_devices,omitempty"`
	LogoutMethod  string   `json:"logout_method,omitempty"`
}

type VerificationEvent struct {
	BaseEvent
	PhoneNumber      string    `json:"phone_number"`
	VerificationCode string    `json:"verification_code"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type NotificationEvent struct {
	BaseEvent
	Message       string   `json:"message"`
	TargetDevices []string `json:"target_devices,omitempty"`
}

type UserActionEvent struct {
	BaseEvent
	Action        string      `json:"action"`
	Details       interface{} `json:"details,omitempty"`
	OriginDevice  string      `json:"origin_device,omitempty"`
	TargetDevices []string    `json:"target_devices,omitempty"`
}

type ResetPasswordEvent struct {
	BaseEvent
	PhoneNumber  string    `json:"phone_number"`
	ResetCode    string    `json:"reset_code"`
	ExpiresAt    time.Time `json:"expires_at"`
	TargetDevice string    `json:"target_device"`
}
