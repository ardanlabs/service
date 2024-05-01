package errs

import (
	"fmt"
)

var (
	// OK indicates the operation was successful.
	OK = ErrCode{value: 0}

	// Canceled indicates the operation was canceled (typically by the caller).
	Canceled = ErrCode{value: 1}

	// Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	Unknown = ErrCode{value: 2}

	// InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	InvalidArgument = ErrCode{value: 3}

	// DeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	DeadlineExceeded = ErrCode{value: 4}

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	NotFound = ErrCode{value: 5}

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists = ErrCode{value: 6}

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	PermissionDenied = ErrCode{value: 7}

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	ResourceExhausted = ErrCode{value: 8}

	// FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	FailedPrecondition = ErrCode{value: 9}

	// Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	Aborted = ErrCode{value: 10}

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
	OutOfRange = ErrCode{value: 11}

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented = ErrCode{value: 12}

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	Internal = ErrCode{value: 13}

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	Unavailable = ErrCode{value: 14}

	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss = ErrCode{value: 15}

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	Unauthenticated = ErrCode{value: 16}
)

// ErrCode represents an error code in the system.
type ErrCode struct {
	value int
}

// Value returns the integer value of the error code.
func (ec ErrCode) Value() int {
	return ec.value
}

// String returns the string representation of the error code.
func (ec ErrCode) String() string {
	return codeNames[ec.value]
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (ec *ErrCode) UnmarshalText(data []byte) error {
	errName := string(data)

	v, exists := codeNumbers[errName]
	if !exists {
		return fmt.Errorf("err code %q does not exist", errName)
	}

	*ec = v

	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (ec ErrCode) MarshalText() ([]byte, error) {
	return []byte(ec.String()), nil
}

// Equal provides support for the go-cmp package and testing.
func (ec ErrCode) Equal(ec2 ErrCode) bool {
	return ec.value == ec2.value
}

var codeNumbers = map[string]ErrCode{
	"ok":                  OK,
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
}

var codeNames [17]string

func init() {
	codeNames[OK.value] = "ok"
	codeNames[Canceled.value] = "canceled"
	codeNames[Unknown.value] = "unknown"
	codeNames[InvalidArgument.value] = "invalid_argument"
	codeNames[DeadlineExceeded.value] = "deadline_exceeded"
	codeNames[NotFound.value] = "not_found"
	codeNames[AlreadyExists.value] = "already_exists"
	codeNames[PermissionDenied.value] = "permission_denied"
	codeNames[ResourceExhausted.value] = "resource_exhausted"
	codeNames[FailedPrecondition.value] = "failed_precondition"
	codeNames[Aborted.value] = "aborted"
	codeNames[OutOfRange.value] = "out_of_range"
	codeNames[Unimplemented.value] = "unimplemented"
	codeNames[Internal.value] = "internal"
	codeNames[Unavailable.value] = "unavailable"
	codeNames[DataLoss.value] = "data_loss"
	codeNames[Unauthenticated.value] = "unauthenticated"
}
