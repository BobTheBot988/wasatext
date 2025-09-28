package database

import (
	"errors"
	"regexp"
	"unicode"
)

func (db *appdbimpl) SanitizeString(s string) (string, error) {

	// Check for alphanumeric and underscore pattern
	alphanumericPattern := `^[a-zA-Z0-9_]*`
	alphanumericMatch, err := regexp.MatchString(alphanumericPattern, s)
	if err != nil {
		return "", err
	}

	// Check for Unicode emojis using Go's unicode package
	hasEmoji := containsEmoji(s)

	// String is valid if it matches any of these patterns
	if alphanumericMatch || hasEmoji {
		return s, nil
	}

	return "", errors.New("invalid string")
}

// containsEmoji checks if string contains emoji characters
func containsEmoji(s string) bool {
	for _, r := range s {
		// Check for emoji ranges
		if isEmoji(r) {
			return true
		}
	}
	return false
}

// isEmoji checks if a rune is an emoji
func isEmoji(r rune) bool {
	// Common emoji ranges
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x1F1E0 && r <= 0x1F1FF) || // Regional indicator symbols
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation selectors
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1F018 && r <= 0x1F270) || // Various symbols
		unicode.Is(unicode.So, r) || // Other symbols
		unicode.Is(unicode.Sk, r) // Symbol modifier
}
