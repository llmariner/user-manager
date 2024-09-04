package userid

import "strings"

// Normalize normalizes the user ID (= email address) to lowercase.
func Normalize(userID string) string {
	return strings.ToLower(userID)
}
