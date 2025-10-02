package utils

import (
	"errors"
	"regexp"
	"strings"
)

// Helper function for custom validation
func ValidateUserInput(fullName, email, phone, password string) error {
    // Validate full name
    if len(strings.TrimSpace(fullName)) < 2 {
        return errors.New("full name must be at least 2 characters")
    }

    // Check for valid characters in full name
    nameRegex := regexp.MustCompile(`^[a-zA-Z\s]+$`)
    if !nameRegex.MatchString(strings.TrimSpace(fullName)) {
        return errors.New("full name can only contain letters and spaces")
    }

    // Validate email format
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return errors.New("invalid email format")
    }

    // Validate phone (Indonesian format)
    phoneRegex := regexp.MustCompile(`^(\+62|62|0)8[1-9][0-9]{6,9}$`)
    if !phoneRegex.MatchString(phone) {
        return errors.New("invalid phone number format (use Indonesian format)")
    }

    // Validate password strength
    if len(password) < 6 {
        return errors.New("password must be at least 6 characters")
    }

    // Check if password contains at least one letter and one number
    hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    if !hasLetter || !hasNumber {
        return errors.New("password must contain at least one letter and one number")
    }

    return nil
}