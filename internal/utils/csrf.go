package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	CSRFTokenMaxAge = 24 * time.Hour
)

func GenerateCSRFToken(userID string, secretKey string) string {
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%s:%d", userID, timestamp)
	
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))

	token := fmt.Sprintf("%s:%s", data, signature)
	return base64.URLEncoding.EncodeToString([]byte(token))
}

func ValidateCSRFToken(token string, expectedUserID string, secretKey string) (bool, error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false, fmt.Errorf("Invalid token format")
	}
	parts := strings.Split(string(decoded), ":")
	if len(parts) != 3 {
		return false, fmt.Errorf("malformed token")
	}

	userID := parts[0]
	timestampStr := parts[1]
	providedSignature := parts[2]

	if userID != expectedUserID {
		return false, fmt.Errorf("token user mismatch")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp")
	}

	tokenTime := time.Unix(timestamp,0)
	if time.Since(tokenTime) > CSRFTokenMaxAge {
		return false, fmt.Errorf("token expired")
	}

	data := fmt.Sprintf("%s:%d", userID, timestamp)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return false, fmt.Errorf("invalid signature")
	}

	return true, nil

}


func GenerateRandomBytes(n int) ([]byte, error){
	b := make([]byte,n)
	_,err:=rand.Read(b)
	return b,err
}