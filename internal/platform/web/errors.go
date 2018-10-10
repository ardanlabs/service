package web

import "net/http"

type HTTPStatuser interface {
	HTTPStatus() int
}

type statusError struct {
	error
	status int
}

func (e statusError) HTTPStatus() int {
	return e.status
}

func ErrorWithStatus(err error, status int) error {
	return statusError{err, status}
}

func StatusFromError(err error) int {
	inf, ok := err.(HTTPStatuser)
	if !ok {
		return http.StatusInternalServerError
	}
	return inf.HTTPStatus()
}
