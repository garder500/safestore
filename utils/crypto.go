package utils

import (
	"crypto/rand"
	"fmt"
)

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
