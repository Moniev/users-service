package utils

import (
	"strconv"
	"users-service/app/models/ent"
)

func GetUserKeys(user *ent.User) []string {
	keys := make([]string, 0)

	idKey := "user:id:" + strconv.Itoa(user.ID)
	keys = append(keys, idKey)

	mailKey := "user:mail:" + user.Mail
	keys = append(keys, mailKey)

	return keys
}
