package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var (
	// ErrInvalidUUID is returned when a UUID is invalid
	ErrInvalidUUID = errors.New("invalid UUID format")
	// ErrInvalidEmail is returned when an email is invalid
	ErrInvalidEmail = errors.New("invalid email format")
	// ErrInvalidUsername is returned when a username is invalid
	ErrInvalidUsername = errors.New("username must be 3-30 characters and contain only letters, numbers, and underscores")
	// ErrInvalidPassword is returned when a password is invalid
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	// ErrEmptyString is returned when a string is empty
	ErrEmptyString = errors.New("field cannot be empty")
	// ErrStringTooLong is returned when a string exceeds max length
	ErrStringTooLong = errors.New("field exceeds maximum length")
	// ErrInvalidPort is returned when a port is invalid
	ErrInvalidPort = errors.New("port must be between 1 and 65535")
	// ErrInvalidName is returned when a name is invalid
	ErrInvalidName = errors.New("name must be at least 2 characters")
)

// ValidateUUID validates a UUID string
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: UUID cannot be empty", ErrInvalidUUID)
	}
	_, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidUUID, id)
	}
	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("%w: email cannot be empty", ErrInvalidEmail)
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}
	return nil
}

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("%w: username cannot be empty", ErrInvalidUsername)
	}

	if len(username) < 3 || len(username) > 30 {
		return fmt.Errorf("%w: %s (length: %d)", ErrInvalidUsername, username, len(username))
	}

	// Only allow letters, numbers, and underscores
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return fmt.Errorf("%w: %s", ErrInvalidUsername, username)
	}

	return nil
}

// ValidatePassword validates a password
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

// ValidateRequired validates that a string is not empty
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s cannot be empty", ErrEmptyString, field)
	}
	return nil
}

// ValidateMaxLength validates that a string doesn't exceed max length
func ValidateMaxLength(field, value string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%w: %s exceeds maximum length of %d", ErrStringTooLong, field, maxLength)
	}
	return nil
}

// ValidateMinLength validates that a string meets minimum length
func ValidateMinLength(field, value string, minLength int) error {
	if len(value) < minLength {
		return fmt.Errorf("%s must be at least %d characters", field, minLength)
	}
	return nil
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

// ValidateName validates a name field
func ValidateName(name string) error {
	if len(strings.TrimSpace(name)) < 2 {
		return fmt.Errorf("%w: %s", ErrInvalidName, name)
	}
	return nil
}

// ValidateSQL validates SQL for basic safety checks
func ValidateSQL(sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return errors.New("SQL query cannot be empty")
	}

	// Check for suspicious patterns (basic injection prevention)
	dangerousPatterns := []string{
		";\\s*DROP",
		";\\s*DELETE\\s+FROM",
		";\\s*TRUNCATE",
		";\\s*ALTER",
		";\\s*CREATE",
		"EXEC\\s*\\(",
		"EXECUTE\\s*IMMEDIATE",
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(strings.ToUpper(sql), pattern)
		if matched {
			return fmt.Errorf("SQL contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateDataSourceType validates a data source type
func ValidateDataSourceType(dsType string) error {
	validTypes := map[string]bool{
		"postgresql": true,
		"mysql":      true,
	}

	if !validTypes[strings.ToLower(dsType)] {
		return fmt.Errorf("invalid data source type: %s (must be 'postgresql' or 'mysql')", dsType)
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	result := strings.Builder{}
	for _, r := range s {
		if !unicode.IsControl(r) || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(page, perPage int) (int, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}
	return page, perPage, nil
}
