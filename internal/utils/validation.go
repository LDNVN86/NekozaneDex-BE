package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// PasswordPolicy - Cấu hình yêu cầu mật khẩu
type PasswordPolicy struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordPolicy - Policy mặc định cho production
var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:        8,
	MaxLength:        128,
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumber:    true,
	RequireSpecial:   false, // Không bắt buộc special char để UX tốt hơn
}

// ValidatePassword - Kiểm tra mật khẩu theo policy
func ValidatePassword(password string, policy PasswordPolicy) error {
	// Check length
	if len(password) < policy.MinLength {
		return errors.New("mật khẩu phải có ít nhất " + string(rune('0'+policy.MinLength)) + " ký tự")
	}
	if len(password) > policy.MaxLength {
		return errors.New("mật khẩu không được quá 128 ký tự")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return errors.New("mật khẩu phải có ít nhất 1 chữ in hoa")
	}
	if policy.RequireLowercase && !hasLower {
		return errors.New("mật khẩu phải có ít nhất 1 chữ thường")
	}
	if policy.RequireNumber && !hasNumber {
		return errors.New("mật khẩu phải có ít nhất 1 số")
	}
	if policy.RequireSpecial && !hasSpecial {
		return errors.New("mật khẩu phải có ít nhất 1 ký tự đặc biệt")
	}

	// Check common weak passwords
	if isCommonPassword(password) {
		return errors.New("mật khẩu quá phổ biến, vui lòng chọn mật khẩu khác")
	}

	return nil
}

// ValidatePasswordDefault - Validate với policy mặc định
func ValidatePasswordDefault(password string) error {
	return ValidatePassword(password, DefaultPasswordPolicy)
}

// isCommonPassword - Kiểm tra mật khẩu phổ biến
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"monkey", "1234567", "letmein", "trustno1", "dragon",
		"baseball", "iloveyou", "master", "sunshine", "ashley",
		"bailey", "shadow", "123123", "654321", "superman",
		"qazwsx", "michael", "football", "password1", "password123",
		"welcome", "welcome1", "admin", "login", "passw0rd",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	return false
}

// ValidateEmail - Kiểm tra email hợp lệ
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername - Kiểm tra username hợp lệ
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("tên người dùng phải có ít nhất 3 ký tự")
	}
	if len(username) > 30 {
		return errors.New("tên người dùng không được quá 30 ký tự")
	}

	// Chỉ cho phép chữ, số và underscore
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !usernameRegex.MatchString(username) {
		return errors.New("tên người dùng chỉ được chứa chữ, số và dấu gạch dưới")
	}

	// Không cho phép bắt đầu bằng số
	if unicode.IsDigit(rune(username[0])) {
		return errors.New("tên người dùng không được bắt đầu bằng số")
	}

	return nil
}

// SanitizeInput - Loại bỏ các ký tự nguy hiểm khỏi input
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Loại bỏ null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	return input
}
