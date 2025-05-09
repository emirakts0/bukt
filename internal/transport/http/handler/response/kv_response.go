package response

import "time"

// todo: time.Time türünde değilde string dönelim ExpiresAt olarak, hem istenen formatta döneriz hemde default değerden kurtuluruz, boş gönderilebilir.

type KVResponse struct {
	Message   string    `json:"message"`
	Key       string    `json:"key,omitempty"`
	Value     string    `json:"value,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
