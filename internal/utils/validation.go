package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

type PasswordPolicy struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:        8,
	MaxLength:        128,
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumber:    true,
	RequireSpecial:   false,
}

func ValidatePassword(password string, policy PasswordPolicy) error {
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

	if isCommonPassword(password) {
		return errors.New("mật khẩu quá phổ biến, vui lòng chọn mật khẩu khác")
	}

	return nil
}

func ValidatePasswordDefault(password string) error {
	return ValidatePassword(password, DefaultPasswordPolicy)
}

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

func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return errors.New("tên người dùng phải có ít nhất 3 ký tự")
	}
	if len(username) > 50 {
		return errors.New("tên người dùng không được quá 50 ký tự")
	}

	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != ' ' {
			return errors.New("tên người dùng chỉ được chứa chữ, số, dấu cách và dấu gạch dưới")
		}
	}

	firstRune := []rune(username)[0]
	if unicode.IsDigit(firstRune) || firstRune == ' ' {
		return errors.New("tên người dùng không được bắt đầu bằng số hoặc dấu cách")
	}

	return nil
}

func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\x00", "")
	return input
}
