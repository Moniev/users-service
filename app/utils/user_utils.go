package utils

import (
	"fmt"
	"users-service/app/models/ent"
)

func GetUserKeys(user *ent.User) []string {
	keys := []string{
		fmt.Sprintf("user:%d", user.ID),
		fmt.Sprintf("user:email:%s", user.Mail),
	}

	return keys
}
