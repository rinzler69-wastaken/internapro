package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateToken(length int) (string, error) {
	if length < 16 {
		length = 16
	}
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
