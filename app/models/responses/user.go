package responses

import "users-service/app/models/ent"

type User struct {
	User  *ent.User `json:"user"`
	Token string    `json:"token,omitempty"`
}
