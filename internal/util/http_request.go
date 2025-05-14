package util

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	// DefaultMaxBodySize is the default maximum size for request bodies (1MB)
	DefaultMaxBodySize = 1 << 20
)

func ReadJSONBodyWithLimit(r *http.Request, v interface{}, maxSize int64) error {
	if maxSize <= 0 {
		maxSize = DefaultMaxBodySize
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxSize))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	return nil
}

func ReadJSONBody(r *http.Request, v interface{}) error {
	return ReadJSONBodyWithLimit(r, v, DefaultMaxBodySize)
}
