package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	numericRe     = regexp.MustCompile(`^[0-9]+$`)
	countryCodeRe = regexp.MustCompile(`^[a-zA-Z]{2}$`)
	dateRe        = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}(-[0-9]{2})?$`)
)

// AppID validates a single App ID (numeric only).
func AppID(id string) error {
	if id == "" {
		return fmt.Errorf("app ID must not be empty")
	}
	if !numericRe.MatchString(id) {
		return fmt.Errorf("invalid app ID %q: must contain only digits", id)
	}
	return nil
}

// AppIDs validates a comma-separated list of App IDs.
func AppIDs(input string) error {
	if input == "" {
		return fmt.Errorf("app IDs must not be empty")
	}
	ids := strings.Split(input, ",")
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if err := AppID(id); err != nil {
			return err
		}
	}
	return nil
}

// CountryCode validates a two-letter country code.
func CountryCode(code string) error {
	if !countryCodeRe.MatchString(code) {
		return fmt.Errorf("invalid country code %q: must be exactly 2 letters", code)
	}
	return nil
}

// DateString validates a date string in YYYY-MM-DD or YYYY-MM format.
func DateString(date string) error {
	if date == "" {
		return fmt.Errorf("date must not be empty")
	}
	if !dateRe.MatchString(date) {
		return fmt.Errorf("invalid date %q: must be YYYY-MM-DD or YYYY-MM", date)
	}
	return nil
}

// FilePath validates a file path for safety.
func FilePath(path string) error {
	if path == "" {
		return fmt.Errorf("file path must not be empty")
	}
	for _, c := range path {
		if c < 0x20 || c == 0x7f {
			return fmt.Errorf("invalid file path: contains control characters")
		}
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid file path: path traversal not allowed")
	}
	return nil
}
