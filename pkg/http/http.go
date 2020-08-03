// Package http contains utility functions for request and response handling.
package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorCode int

const (
	ErrorCodeRoomNotFound ErrorCode = iota
	ErrorCodeRoomFailedToJoin
	ErrorCodeInvalidRequestBody
	ErrorCodeFailedToCreateRoom
	ErrorCodeMissingParameter
	ErrorCodeFailedToRegister
	ErrorCodeInvalidEmail
	ErrorCodeInvalidUsername
	ErrorCodeUsernameAlreadyExists
	ErrorCodeFailedToLogin
	ErrorCodeUnauthorized
)

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

// JsonEncode marshals an interface and writes it to the response.
func JsonEncode(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
