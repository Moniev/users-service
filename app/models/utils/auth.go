package utils

import (
	"users-service/app/models/ent"

	"github.com/golang-jwt/jwt/v5"
)

// PermissionInfo represents a single permission with its associated capabilities.
type PermissionInfo struct {
	Name   string `json:"name"`   // The identifier of the permission.
	View   bool   `json:"view"`   // Indicates whether viewing is allowed.
	Add    bool   `json:"add"`    // Indicates whether adding is allowed.
	Edit   bool   `json:"edit"`   // Indicates whether editing is allowed.
	Delete bool   `json:"delete"` // Indicates whether deletion is allowed.
}

// UserRoleInfo represents a user role with its description and associated permissions.
type UserRoleInfo struct {
	Name        string           `json:"name"`        // The identifier of the role.
	Description string           `json:"description"` // Details about the role's purpose.
	Permissions []PermissionInfo `json:"permissions"` // Slice of permissions defining role capabilities.
}

// Claims defines the JWT claims structure for user authentication.
// It includes user ID and roles, along with standard JWT registered claims (e.g., expiration).
type Claims struct {
	UserID               int               `json:"user_id"` // The user’s unique identifier.
	DeviceID             int               `json:"device_id"`
	UserRoles            []UserRoleInfo    `json:"user_roles"` // List of user roles (e.g., "admin", "user").
	jwt.RegisteredClaims `json:"-,inline"` // Standard JWT fields like ExpiresAt, IssuedAt.
}

// JWTValidationResult represents the outcome of a JWT validation process.
// It contains extracted claims and any potential validation errors.
type JWTValidationResult struct {
	Claims *Claims `json:"claims"` // Parsed claims from the JWT token if validation succeeds
	Err    error   `json:"error"`  // Error encountered during validation, if any
}

// Jwt result is struct attending results of jwt hashing threads.
// It contains variable token and potential error.
type JWTResult struct {
	User  *ent.User `json:"user"`
	Token string    `json:"token"` //String of the token
	Err   error     `json:"error"` // Potential error
}

// Hash result is struct attending results of hashing threads.
// It contains variable bool and potential error.
type HashResult struct {
	Valid bool  `json:"valid"` //Result of validation
	Err   error `json:"error"` // Potential error
}
