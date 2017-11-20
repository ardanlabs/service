package web

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound is abstracting the mgo not found error.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in it's proper form")

	// ErrDBNotConfigured occurs when the DB is not initialized.
	ErrDBNotConfigured = errors.New("DB not initialized")
)

// JSONError is the response for errors that occur within the API.
type JSONError struct {
	Error string `json:"error"`
}

// Error handles all error responses for the API.
func Error(cxt context.Context, w http.ResponseWriter, err error) {
	switch errors.Cause(err) {
	case ErrNotFound:
		RespondError(cxt, w, err, http.StatusNotFound)
		return

	case ErrInvalidID:
		RespondError(cxt, w, err, http.StatusBadRequest)
		return
	}

	RespondError(cxt, w, err, http.StatusInternalServerError)
}

// RespondError sends JSON describing the error
func RespondError(ctx context.Context, w http.ResponseWriter, err error, code int) {
	Respond(ctx, w, JSONError{Error: err.Error()}, code)
}

// Respond sends JSON to the client.
// If code is StatusNoContent, v is expected to be nil.
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, code int) {

	// Set the status code for the request logger middleware.
	v := ctx.Value(KeyValues).(*Values)
	v.StatusCode = code

	// Just set the status code and we are done.
	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return
	}

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response and context.
	w.WriteHeader(code)

	// Marshal the data into a JSON string.
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("%s : Respond %v Marshalling JSON response\n", v.TraceID, err)
		jsonData = []byte("{}")
	}

	// Send the result back to the client.
	io.WriteString(w, string(jsonData))
}
