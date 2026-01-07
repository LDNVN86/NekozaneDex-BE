package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken - Hash token bằng SHA256
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
