package utils

import "strings"

const (
	whiteListChar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

// IsGlobalServerIDPermitted validates if the serverID matches the required format and rules.
func IsGlobalServerIDPermitted(serverID string) bool {
	// 1. Split the serverID into prefix and exampleid.
	// SplitN ensures we only split at the first ".s" in case exampleid contains ".s".
	parts := strings.SplitN(serverID, ".s.", 2)
	if len(parts) != 2 {
		return false // ".s" separator not found or improperly formatted
	}

	prefix := parts[0]
	suffixID := parts[1]

	// 2. Prefix cannot be empty
	if len(prefix) == 0 {
		return false
	}

	// 3. Validate the starting character of the prefix
	// Allowed: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" (Notice 'Z' is excluded)
	firstChar := prefix[0]
	isValidStart := (firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')
	if !isValidStart {
		return false
	}

	// 4. Validate all characters in the prefix
	// Allowed: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		isValidChar := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isValidChar {
			return false
		}
	}

	// 5. Pass the exampleid to your existing validation function
	return IsSuffixIDPermited(suffixID)
}

// IsGlobalClientIDPermitted validates if the serverID matches the required format and rules.
func IsGlobalClientIDPermitted(clientID string) bool {
	// 1. Split the serverID into prefix and exampleid.
	// SplitN ensures we only split at the first ".s" in case exampleid contains ".s".
	parts := strings.SplitN(clientID, ".c.", 2)
	if len(parts) != 2 {
		return false // ".s" separator not found or improperly formatted
	}

	prefix := parts[0]
	suffixID := parts[1]

	// 2. Prefix cannot be empty
	if len(prefix) == 0 {
		return false
	}

	// 3. Validate the starting character of the prefix
	// Allowed: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	firstChar := prefix[0]
	isValidStart := (firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')
	if !isValidStart {
		return false
	}

	// 4. Validate all characters in the prefix
	// Allowed: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < len(prefix); i++ {
		c := prefix[i]
		isValidChar := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isValidChar {
			return false
		}
	}

	// 5. Pass the exampleid to your existing validation function
	return IsSuffixIDPermited(suffixID)
}

func IsSuffixIDPermited(clientID string) bool {
	if len(clientID) == 0 {
		return false
	}

	chrMap := make(map[rune]bool)
	for _, chr := range whiteListChar {
		chrMap[chr] = true
	}

	for _, chr := range clientID {
		if !chrMap[chr] {
			return false
		}
	}

	return true
}

func MakeClientIDPermited(clientID string) string {
	input := []rune(clientID)
	output := input
	chrMap := make(map[rune]bool)
	for _, chr := range whiteListChar {
		chrMap[chr] = true
	}
	for idx, chr := range input {
		if !chrMap[chr] {
			output[idx] = '-'
		}
	}
	return string(output)
}
