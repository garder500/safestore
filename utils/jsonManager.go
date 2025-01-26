package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

func JsonError(title, message string, code int) ([]byte, error) {
	errorMessage := map[string]interface{}{
		"error": map[string]interface{}{
			"title":   title,
			"message": message,
			"code":    code,
		},
	}
	return formatJsonToResponse(errorMessage)
}

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func formatJsonToResponse(data interface{}) ([]byte, error) {
	formated, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return formated, nil
}

func FormatHttpError(w http.ResponseWriter, httpCode int, title, message string) {
	data, err := JsonError(title, message, httpCode)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	http.Error(w, string(data), httpCode)
}

func FormatHttpSuccess(w http.ResponseWriter, data interface{}) {
	formated, err := formatJsonToResponse(data)
	if err != nil {
		FormatHttpError(w, 500, "Internal server error", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(formated)
}
