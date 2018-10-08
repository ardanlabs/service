package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

var (
	// ErrNotHealthy occurs when the service is having problems.
	ErrNotHealthy = errors.New("Not healthy")

	// ErrNotFound is abstracting the mgo not found error.
	ErrNotFound = errors.New("Entity not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrValidation occurs when there are validation errors.
	ErrValidation = errors.New("Validation errors occurred")

	// ErrUnauthorized occurs when there was an issue validing the client's
	// credentials.
	ErrUnauthorized = errors.New("Unauthorized")

	// ErrForbidden occurs when we know who the user is but they attempt a
	// forbidden action.
	ErrForbidden = errors.New("Forbidden")
)

// JSONError is the response for errors that occur within the API.
type JSONError struct {
	Error  string       `json:"error"`
	Fields InvalidError `json:"fields,omitempty"`
}

// Error handles all error responses for the API.
func Error(cxt context.Context, log *log.Logger, w http.ResponseWriter, err error) {
	switch errors.Cause(err) {
	case ErrNotHealthy:
		RespondError(cxt, log, w, err, http.StatusInternalServerError)
		return

	case ErrNotFound:
		RespondError(cxt, log, w, err, http.StatusNotFound)
		return

	case ErrValidation, ErrInvalidID:
		RespondError(cxt, log, w, err, http.StatusBadRequest)
		return

	case ErrUnauthorized:
		RespondError(cxt, log, w, err, http.StatusUnauthorized)
		return

	case ErrForbidden:
		RespondError(cxt, log, w, err, http.StatusForbidden)
		return
	}

	switch e := errors.Cause(err).(type) {
	case InvalidError:
		v := JSONError{
			Error:  "field validation failure",
			Fields: e,
		}

		Respond(cxt, log, w, v, http.StatusBadRequest)
		return
	}

	RespondError(cxt, log, w, err, http.StatusInternalServerError)
}

// RespondError sends JSON describing the error
func RespondError(ctx context.Context, log *log.Logger, w http.ResponseWriter, err error, code int) {
	Respond(ctx, log, w, JSONError{Error: err.Error()}, code)
}

// Respond sends JSON to the client.
// If code is StatusNoContent, v is expected to be nil.
func Respond(ctx context.Context, log *log.Logger, w http.ResponseWriter, data interface{}, code int) {

	// Set the status code for the request logger middleware.
	v := ctx.Value(KeyValues).(*Values)
	v.StatusCode = code

	// Just set the status code and we are done. If there is nothing to marshal
	// set status code and return.
	if code == http.StatusNoContent || data == nil {
		w.WriteHeader(code)
		return
	}

	// Marshal the data into a JSON string.
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("%s : Respond %v Marshalling JSON response\n", v.TraceID, err)

		// Should respond with internal server error.
		RespondError(ctx, log, w, err, http.StatusInternalServerError)
		return
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response and context.
	w.WriteHeader(code)

	// Send the result back to the client.
	w.Write(jsonData)
}
