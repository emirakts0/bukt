package http_util

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		_, writeErr := w.Write([]byte(`{"error":"failed to encode JSON"}`))
		if writeErr != nil {
			slog.Error("Failed to write error response", "error", writeErr)
		}
		return
	}
}

func JSONError(w http.ResponseWriter, status int, errorData any) {
	errorReponse := map[string]any{
		"error":     errorData,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	JSON(w, status, errorReponse)
}

// WriteCreated writes a 201 Created response with the given data
func WriteCreated(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// WriteOK writes a 200 response with the given data
func WriteOK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// WriteBadRequest writes a 400 response with the given message
func WriteBadRequest(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusBadRequest, message)
}

// WriteNotFound writes a 404 response with the given message
func WriteNotFound(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusNotFound, message)
}

// WriteInternalError writes a 500 response with the given message
func WriteInternalError(w http.ResponseWriter) {
	JSONError(w, http.StatusInternalServerError, "Internal server error")
}

// WriteUnauthorized writes a 401 response with the given message
func WriteUnauthorized(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusUnauthorized, message)
}

// WriteNoContent writes a 204 response with the given message
func WriteNoContent(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNoContent, message)
}
