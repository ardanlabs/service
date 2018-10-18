package web

import (
	"fmt"
	"net/http"
)

type statusError struct {
	err    error
	Status int
}

func (se *statusError) Error() string {
	return fmt.Sprintf("http status %v: %s", se.Status, se.err)
}

func ErrorWithStatus(err error, status int) error {
	return &statusError{err, status}
}

func StatusFromError(err error) int {
	serr, ok := err.(*statusError)
	if !ok {
		return http.StatusInternalServerError
	}
	return serr.Status
}
