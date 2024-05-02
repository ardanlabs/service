package response

import (
	"errors"
	"net/http"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/api/mid"
	"github.com/ardanlabs/service/foundation/web"
)

var httpStatus = make(map[errs.ErrCode]int)
var errsStatus = make(map[int]errs.ErrCode)

// init maps out the error codes to http status codes.
func init() {
	httpStatus[errs.OK] = http.StatusOK
	httpStatus[errs.NoContent] = http.StatusNoContent
	httpStatus[errs.StatusCreated] = http.StatusCreated
	httpStatus[errs.Canceled] = http.StatusGatewayTimeout
	httpStatus[errs.Unknown] = http.StatusInternalServerError
	httpStatus[errs.InvalidArgument] = http.StatusBadRequest
	httpStatus[errs.DeadlineExceeded] = http.StatusGatewayTimeout
	httpStatus[errs.NotFound] = http.StatusNotFound
	httpStatus[errs.AlreadyExists] = http.StatusConflict
	httpStatus[errs.PermissionDenied] = http.StatusForbidden
	httpStatus[errs.ResourceExhausted] = http.StatusTooManyRequests
	httpStatus[errs.FailedPrecondition] = http.StatusBadRequest
	httpStatus[errs.Aborted] = http.StatusConflict
	httpStatus[errs.OutOfRange] = http.StatusBadRequest
	httpStatus[errs.Unimplemented] = http.StatusNotImplemented
	httpStatus[errs.Internal] = http.StatusInternalServerError
	httpStatus[errs.Unavailable] = http.StatusServiceUnavailable
	httpStatus[errs.DataLoss] = http.StatusInternalServerError
	httpStatus[errs.Unauthenticated] = http.StatusUnauthorized

	errsStatus[http.StatusOK] = errs.OK
	errsStatus[http.StatusNoContent] = errs.NoContent
	errsStatus[http.StatusCreated] = errs.StatusCreated
	errsStatus[http.StatusGatewayTimeout] = errs.DeadlineExceeded
	errsStatus[http.StatusNotFound] = errs.NotFound
	errsStatus[http.StatusForbidden] = errs.PermissionDenied
	errsStatus[http.StatusTooManyRequests] = errs.ResourceExhausted
	errsStatus[http.StatusBadRequest] = errs.FailedPrecondition
	errsStatus[http.StatusConflict] = errs.Aborted
	errsStatus[http.StatusNotImplemented] = errs.Unimplemented
	errsStatus[http.StatusInternalServerError] = errs.Internal
	errsStatus[http.StatusServiceUnavailable] = errs.Unavailable
	errsStatus[http.StatusUnauthorized] = errs.Unauthenticated
}

func ToMid(resp web.Response) mid.Response {
	var err *errs.Error
	errors.As(resp.Err, &err)

	return mid.Response{
		Errs:       err,
		Data:       resp.Data,
		StatusCode: errsStatus[resp.StatusCode],
	}
}

func ToWebX(name string, resp mid.Response) web.Response {
	var err error
	if resp.Errs != nil {
		err = resp.Errs
	}

	return web.Response{
		Err:        err,
		Data:       resp.Data,
		StatusCode: httpStatus[resp.StatusCode],
	}
}

func ToWeb(resp mid.Response) web.Response {
	var err error
	if resp.Errs != nil {
		err = resp.Errs
	}

	return web.Response{
		Err:        err,
		Data:       resp.Data,
		StatusCode: httpStatus[resp.StatusCode],
	}
}

func Response(data any, httpStatusCode int) web.Response {
	return web.Respond(data, httpStatusCode)
}

func AppErrs(err *errs.Error) web.Response {
	if err != nil {
		return web.RespondError(err, httpStatus[err.Code])
	}

	return web.EmptyResponse()
}

func AppError(code errs.ErrCode, err error) web.Response {
	return web.RespondError(errs.New(code, err), httpStatus[code])
}

func AppAPIError(err error) web.Response {
	var ers *errs.Error
	if !errors.As(err, &ers) {
		ers = errs.New(errs.Internal, err)
	}

	return web.RespondError(ers, httpStatus[ers.Code])
}

func AppErrorf(code errs.ErrCode, format string, v ...any) web.Response {
	return web.RespondError(errs.Newf(code, format, v...), httpStatus[code])
}
