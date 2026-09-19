package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"jian-unified-system/apollo/apollo-api/internal/application"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Parse bounds authentication payloads before go-zero maps them to generated types.
func Parse(w http.ResponseWriter, r *http.Request, v any) error {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<10))
	if err != nil || !json.Valid(body) {
		return application.ErrInvalid
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err = httpx.Parse(r, v); err != nil {
		return application.ErrInvalid
	}
	return nil
}
