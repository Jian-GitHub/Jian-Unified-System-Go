package util

import (
	"encoding/base64"
	"encoding/json"
)

// Base64EncodeToString encodes any Go value as JSON, then returns its base64 URL-encoded string.
func Base64EncodeToString(v any) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(jsonBytes)
	return encoded, nil
}

// Base64Decode decodes a base64 URL-encoded string and returns a slice of byte.
func base64Decode(encoded string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(encoded)
}
