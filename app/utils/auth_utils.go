package utils

import (
	"encoding/json"
	"users-service/app/models/ent"
	"users-service/app/models/utils"

	"github.com/gin-gonic/gin"
)

// GetUserID retrieves the user ID from the Gin context.
// It returns the user ID as an integer if present and valid, or 0 if not found or invalid.
// The user ID is expected to be stored in the context under the "UserID" key by authentication middleware.
func GetUserID(ctx *gin.Context) int {
	userIDRaw, exists := ctx.Get("UserID")
	if !exists {
		return 0
	}

	if userID, ok := userIDRaw.(int); ok {
		return userID
	}

	if userIDFloat, ok := userIDRaw.(float64); ok {
		return int(userIDFloat)
	}

	return 0
}

// BindRoles retrieves user roles from the Gin context.
// It returns a slice of UserRoleInfo if roles are present and valid, or an empty slice if not found or invalid.
// The roles are expected to be stored in the context under the "Roles" key by authentication middleware.
func BindRoles(ctx *gin.Context) []utils.UserRoleInfo { // Use alias for UserRoleInfo
	rolesRaw, exists := ctx.Get("UserRoles")
	if !exists {
		return []utils.UserRoleInfo{}
	}

	if roles, ok := rolesRaw.([]utils.UserRoleInfo); ok {
		return roles
	}

	if rolesInterfaceSlice, ok := rolesRaw.([]interface{}); ok {
		marshaledRoles, err := json.Marshal(rolesInterfaceSlice)
		if err != nil {
			return []utils.UserRoleInfo{}
		}

		var roles []utils.UserRoleInfo
		err = json.Unmarshal(marshaledRoles, &roles)
		if err != nil {
			return []utils.UserRoleInfo{}
		}

		return roles
	}

	return []utils.UserRoleInfo{}
}

func ValidateRoles(roles []utils.UserRoleInfo, names []string) bool {
	for _, role := range roles {
		if Contains(names, role.Name) {
			return true
		}
	}

	return false
}

// GetToken retrieves the authentication token from the Gin context.
// It returns the token as a string if present and valid, or an empty string if not found or invalid.
// The token is expected to be stored in the context under the "Token" key by authentication middleware.
func GetToken(ctx *gin.Context) string {
	tokenRaw, exists := ctx.Get("Token")
	if !exists {
		return ""
	}

	token, ok := tokenRaw.(string)
	if ok {
		return token
	}

	return ""
}

// MarshalUserRoles takes a slice of UserRole entities and converts them into a slice of UserRoleInfo structs.
// It iterates over each UserRole, extracting the relevant permissions and populating a corresponding PermissionInfo struct.
// Each UserRoleInfo contains the name, description, and a slice of permissions associated with that role.
// The function returns a slice of UserRoleInfo structs that can be used for further processing or serialization.
//
// Parameters:
// - roles: A slice of pointers to UserRole entities. Each UserRole is expected to have associated Permissions.
//
// Returns:
// - A slice of UserRoleInfo structs, where each struct contains the name, description, and permissions for a user role.
func MarshalUserRoles(roles []*ent.UserRole) []utils.UserRoleInfo {
	userRoles := make([]utils.UserRoleInfo, 0, len(roles))
	for _, userRole := range roles {
		permissions := make([]utils.PermissionInfo, 0, len(userRole.Edges.Permissions))

		for _, perm := range userRole.Edges.Permissions {
			permissions = append(permissions, utils.PermissionInfo{
				Name:   perm.Name,
				View:   perm.View,
				Add:    perm.Add,
				Edit:   perm.Edit,
				Delete: perm.Delete,
			})
		}

		userRoles = append(userRoles, utils.UserRoleInfo{
			Name:        userRole.Name,
			Description: userRole.Description,
			Permissions: permissions,
		})
	}

	return userRoles
}
