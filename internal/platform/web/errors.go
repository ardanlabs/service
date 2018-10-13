package web

import (
	"net/http"
)

type statusError struct {
	error
	Status int
}

func ErrorWithStatus(err error, status int) error {
	return statusError{err, status}
}

func StatusFromError(err error) int {
	serr, ok := err.(statusError)
	if !ok {
		return http.StatusInternalServerError
	}
	return serr.Status
}
