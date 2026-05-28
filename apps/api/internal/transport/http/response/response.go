// Package response writes the application's uniform JSON envelope.
//
// Every successful response is `{"data": <payload>}`; every error is
// `{"error": {"code": "...", "message": "..."}}` (plus optional fields).
// The shape stays stable so the front-end can have one decoder and one
// error renderer for the whole API.
package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes payload as `{"data": payload}` with the given status.
// Payload is omitted (no "data" key) when nil — handy for 204-like
// shapes where the client doesn't need a body but the envelope would
// look noisy.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(struct {
		Data any `json:"data"`
	}{Data: payload})
}

// Empty writes only the status — no body, no Content-Type. Use for
// responses where every byte is meaningful overhead (e.g. logout 204).
func Empty(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

// ErrorBody is the shape inside the outer `{"error": ...}` envelope.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Error writes the failure envelope at the given status.
func Error(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error ErrorBody `json:"error"`
	}{Error: ErrorBody{Code: code, Message: message, Details: details}})
}
