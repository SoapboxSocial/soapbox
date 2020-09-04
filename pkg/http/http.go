// Package http contains utility functions for request and response handling.
package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorCode int

const (
	ErrorCodeInvalidRequestBody ErrorCode = iota
	ErrorCodeMissingParameter
	ErrorCodeFailedToRegister
	ErrorCodeInvalidEmail
	ErrorCodeInvalidUsername
	ErrorCodeUsernameAlreadyExists
	ErrorCodeFailedToLogin
	ErrorCodeIncorrectPin
	ErrorCodeUserNotFound
	ErrorCodeFailedToGetUser
	ErrorCodeFailedToGetFollowers
	ErrorCodeUnauthorized
	ErrorCodeFailedToStoreDevice
	ErrorCodeNotFound
	ErrorCodeNotAllowed
)

// NotFoundHandler handles 404 responses
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	JsonError(w, http.StatusNotFound, ErrorCodeNotFound, "not found")
}

// NotAllowed handles 405 responses
func NotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	JsonError(w, http.StatusMethodNotAllowed, ErrorCodeNotAllowed, "not allowed")
}

// JsonError writes an Error to the ResponseWriter with the provided information.
func JsonError(w http.ResponseWriter, responseCode int, code ErrorCode, msg string) {
	type ErrorResponse struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	}

	w.WriteHeader(responseCode)

	err := JsonEncode(w, ErrorResponse{Code: code, Message: msg})
	if err != nil {
		log.Printf("failed to encode response: %s", err.Error())
	}
}

// JsonSuccess writes a success message to the writer.
func JsonSuccess(w http.ResponseWriter) {
	type SuccessResponse struct {
		Success bool `json:"success"`
	}

	w.WriteHeader(200)
	err := JsonEncode(w, SuccessResponse{Success: true})
	if err != nil {
		log.Printf("failed to encode response: %s", err.Error())
	}
}

// JsonEncode marshals an interface and writes it to the response.
func JsonEncode(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
