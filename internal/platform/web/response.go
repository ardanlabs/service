package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// RespondErrorMsg wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func RespondErrorMsg(msg string, status int) error {
	return &Error{errors.New(msg), status, nil}
}

// RespondError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func RespondError(err error, status int) error {
	return &Error{err, status, nil}
}

// Respond converts a Go value to JSON and sends it to the client.
// If code is StatusNoContent, v is expected to be nil.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, code int) error {

	// Set the status code for the request logger middleware.
	// If the context is missing this value, request the service
	// to be shutdown gracefully.
	v, ok := ctx.Value(KeyValues).(*Values)
	if !ok {
		return Shutdown("web value missing from context")
	}
	v.StatusCode = code

	// If there is nothing to marshal then set status code and return.
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(code)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
