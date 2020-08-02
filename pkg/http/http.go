// Package http contains utility functions for request and response handling.
package http

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorCode int

const (
	ErrorCodeRoomNotFound       ErrorCode = 1
	ErrorCodeRoomFailedToJoin             = 2
	ErrorCodeInvalidRequestBody           = 3
	ErrorCodeFailedToCreateRoom           = 4
)

// JsonError writes an Error to the ResponseWriter with the provided information.
func JsonError(w http.ResponseWriter, responseCode int, code ErrorCode, msg string) {
	type ErrorResponse struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	}

	resp, err := json.Marshal(ErrorResponse{Code: code, Message: msg})
	if err != nil {
		log.Println("failed encoding error")
		return
	}

	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("failed to encode response: %s", err.Error())
	}
}

// JsonEncode marshals an interface and writes it to the response.
func JsonEncode(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
