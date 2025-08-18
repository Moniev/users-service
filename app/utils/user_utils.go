package utils

import (
	"fmt"
	"users-service/app/models/ent"
)

func GetUserKeys(user *ent.User) []string {
	if user.Edges.UserDetails == nil {
		return []string{fmt.Sprintf("user-public:%d", user.ID)}
	}

	keys := []string{
		fmt.Sprintf("user:%d", user.ID),
		fmt.Sprintf("user:email:%s", user.Mail),
	}

	return keys
}

func RemoveIDs(subIDs, remIDs []int) []int {
	remMap := make(map[int]bool)
	for _, remID := range remIDs {
		remMap[remID] = true
	}

	result := make([]int, 0, len(subIDs))

	for _, id := range subIDs {
		if !remMap[id] {
			result = append(result, id)
		}
	}

	return result
}
