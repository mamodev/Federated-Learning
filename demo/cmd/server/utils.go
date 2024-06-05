package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

func generateRandomKey(length int) (string, error) {
	// Determine the number of random bytes needed
	numBytes := length / 2 // Since each byte will be represented by 2 hex characters

	// Generate random bytes
	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
			return "", err
	}

	// Convert bytes to hexadecimal representation
	apiKey := hex.EncodeToString(bytes)

	return apiKey, nil
}

type Json map[string]interface{}

func (j *Json) String() string {
	b, _ := json.Marshal(j)
	return string(b)
}

func (j *Json) Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(j)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}