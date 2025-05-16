package http_util

import (
	"encoding/json"
	"net/http"
)

const (
	// DefaultMaxBodySize is the default maximum size for request bodies (1MB)
	DefaultMaxBodySize = 1 << 20
)

func ReadJSONBodyWithLimit(r *http.Request, v interface{}, maxSize int64, w http.ResponseWriter) error {
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return err
	}

	return nil
}

func ReadJSONBody(r *http.Request, v interface{}, w http.ResponseWriter) error {
	return ReadJSONBodyWithLimit(r, v, DefaultMaxBodySize, w)
}
