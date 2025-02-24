package errs

import (
	"net/http"
)

var (
	// OK indicates the operation was successful.
	OK = ErrCode{value: 0}

	// NoContent indicates the operation was successful with no content.
	NoContent = ErrCode{value: 1}

	// Canceled indicates the operation was canceled (typically by the caller).
	Canceled = ErrCode{value: 2}

	// Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	Unknown = ErrCode{value: 3}

	// InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	InvalidArgument = ErrCode{value: 4}

	// DeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	DeadlineExceeded = ErrCode{value: 5}

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	NotFound = ErrCode{value: 6}

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists = ErrCode{value: 7}

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	PermissionDenied = ErrCode{value: 8}

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	ResourceExhausted = ErrCode{value: 9}

	// FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	FailedPrecondition = ErrCode{value: 10}

	// Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	Aborted = ErrCode{value: 11}

	// OutOfRange means operation was attempted past the valid range.
	// E.g., seeking or reading past end of file.
	//
	// Unlike InvalidArgument, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate InvalidArgument if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// OutOfRange if asked to read from an offset past the current
	// file size.
	//
	// There is a fair bit of overlap between FailedPrecondition and
	// OutOfRange. We recommend using OutOfRange (the more specific
	// error) when it applies so that callers who are iterating through
	// a space can easily look for an OutOfRange error to detect when
	// they are done.
	OutOfRange = ErrCode{value: 12}

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented = ErrCode{value: 13}

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	Internal = ErrCode{value: 14}

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	Unavailable = ErrCode{value: 15}

	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss = ErrCode{value: 16}

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	Unauthenticated = ErrCode{value: 17}

	// TooManyRequests indicates that the client has made too many requests and
	// exceeded their rate limit and/or quota and must wait before making
	// futhur requests.
	TooManyRequests = ErrCode{value: 18}

	// InternalOnlyLog errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken. The error message is not sent to the client.
	InternalOnlyLog = ErrCode{value: 19}
)

var codeNumbers = map[string]ErrCode{
	"ok":                  OK,
	"no_content":          NoContent,
	"canceled":            Canceled,
	"unknown":             Unknown,
	"invalid_argument":    InvalidArgument,
	"deadline_exceeded":   DeadlineExceeded,
	"not_found":           NotFound,
	"already_exists":      AlreadyExists,
	"permission_denied":   PermissionDenied,
	"resource_exhausted":  ResourceExhausted,
	"failed_precondition": FailedPrecondition,
	"aborted":             Aborted,
	"out_of_range":        OutOfRange,
	"unimplemented":       Unimplemented,
	"internal":            Internal,
	"unavailable":         Unavailable,
	"data_loss":           DataLoss,
	"unauthenticated":     Unauthenticated,
	"too_many_requests":   TooManyRequests,
	"internal_only_log":   InternalOnlyLog,
}

var codeNames = map[ErrCode]string{
	OK:                 "ok",
	NoContent:          "ok_no_content",
	Canceled:           "canceled",
	Unknown:            "unknown",
	InvalidArgument:    "invalid_argument",
	DeadlineExceeded:   "deadline_exceeded",
	NotFound:           "not_found",
	AlreadyExists:      "already_exists",
	PermissionDenied:   "permission_denied",
	ResourceExhausted:  "resource_exhausted",
	FailedPrecondition: "failed_precondition",
	Aborted:            "aborted",
	OutOfRange:         "out_of_range",
	Unimplemented:      "unimplemented",
	Internal:           "internal",
	Unavailable:        "unavailable",
	DataLoss:           "data_loss",
	Unauthenticated:    "unauthenticated",
	TooManyRequests:    "too_many_requests",
	InternalOnlyLog:    "internal_only_log",
}

var httpStatus = map[ErrCode]int{
	OK:                 http.StatusOK,
	NoContent:          http.StatusNoContent,
	Canceled:           http.StatusGatewayTimeout,
	Unknown:            http.StatusInternalServerError,
	InvalidArgument:    http.StatusBadRequest,
	DeadlineExceeded:   http.StatusGatewayTimeout,
	NotFound:           http.StatusNotFound,
	AlreadyExists:      http.StatusConflict,
	PermissionDenied:   http.StatusForbidden,
	ResourceExhausted:  http.StatusTooManyRequests,
	FailedPrecondition: http.StatusBadRequest,
	Aborted:            http.StatusConflict,
	OutOfRange:         http.StatusBadRequest,
	Unimplemented:      http.StatusNotImplemented,
	Internal:           http.StatusInternalServerError,
	Unavailable:        http.StatusServiceUnavailable,
	DataLoss:           http.StatusInternalServerError,
	Unauthenticated:    http.StatusUnauthorized,
	TooManyRequests:    http.StatusTooManyRequests,
	InternalOnlyLog:    http.StatusInternalServerError,
}
