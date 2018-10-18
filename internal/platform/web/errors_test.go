package web_test

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/pkg/errors"
)

func TestStatusError(t *testing.T) {
	cases := []struct {
		Err            error
		ExpectedStatus int
		ExpectedString string
	}{
		{
			Err:            errors.New("utoh"),
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedString: "utoh",
		},
		{
			Err:            web.ErrorWithStatus(errors.New("its not my fault!"), http.StatusBadRequest),
			ExpectedStatus: http.StatusBadRequest,
			ExpectedString: "http status 400: its not my fault!",
		},
		{
			// NOTE: If we wrap the error, we lose the status.
			// TODO: Is this the desired behavior?
			Err:            errors.Wrap(web.ErrorWithStatus(errors.New("its not my fault!"), http.StatusBadRequest), "more info"),
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedString: "more info: http status 400: its not my fault!",
		},
	}

	for i, c := range cases {
		s := web.StatusFromError(c.Err)
		if exp, got := c.ExpectedStatus, s; exp != got {
			t.Fatalf("[%v] expected status %v, got %v", i, exp, got)
		}

		if exp, got := c.ExpectedString, c.Err.Error(); exp != got {
			t.Fatalf("[%v] expected error string %q, got %q", i, exp, got)
		}
	}
}
