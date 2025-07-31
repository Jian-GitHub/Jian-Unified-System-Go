package passkeyUtil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
)

// CreateCredentialRequest JSON -> http.Request
func CreateCredentialRequest(jsonData []byte) (*http.Request, error) {
	req := httptest.NewRequest("POST", "/", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
