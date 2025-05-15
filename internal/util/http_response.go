package util

import (
	"encoding/json"
	"net/http"
	"time"
)

// WriteJSON writes a JSON response with the given status code and data
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	bytes, err := json.Marshal(data)
	if err != nil {
		WriteInternalError(w, "Internal server error")
	}
	if _, err := w.Write(bytes); err != nil {
		WriteInternalError(w, "Internal server error")
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	errorReponse := map[string]string{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	WriteJSON(w, status, errorReponse)
}

// WriteCreated writes a 201 Created response with the given data
func WriteCreated(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// WriteOK writes a 200 OK response with the given data
func WriteOK(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

// WriteBadRequest writes a 400 Bad Request response with the given message
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

// WriteNotFound writes a 404 Not Found response with the given message
func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message)
}

// WriteInternalError writes a 500 Internal Server Error response with the given message
func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}

// WriteUnauthorized writes a 401 Unauthorized response with the given message
func WriteUnauthorized(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusUnauthorized, message)
}
