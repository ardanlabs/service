// Package errs provides types and support related to web error functionality.
package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
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
	return codeNames[ec]
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

// =============================================================================

// Error represents an error in the system.
type Error struct {
	Code     ErrCode `json:"code"`
	Message  string  `json:"message"`
	FuncName string  `json:"-"`
	FileName string  `json:"-"`
}

// New constructs an error based on an app error.
func New(code ErrCode, err error) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	return &Error{
		Code:     code,
		Message:  err.Error(),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// Newf constructs an error based on a error message.
func Newf(code ErrCode, format string, v ...any) *Error {
	pc, filename, line, _ := runtime.Caller(1)

	return &Error{
		Code:     code,
		Message:  fmt.Sprintf(format, v...),
		FuncName: runtime.FuncForPC(pc).Name(),
		FileName: fmt.Sprintf("%s:%d", filename, line),
	}
}

// NewError checks for an Error in the error interface value. If it doesn't
// exist, will create one from the error.
func NewError(err error) *Error {
	var errsErr *Error
	if errors.As(err, &errsErr) {
		return errsErr
	}

	return New(Internal, err)
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Encode implements the encoder interface.
func (e *Error) Encode() ([]byte, string, error) {
	data, err := json.Marshal(e)
	return data, "application/json", err
}

// HTTPStatus implements the web package httpStatus interface so the
// web framework can use the correct http status.
func (e *Error) HTTPStatus() int {
	return httpStatus[e.Code]
}

// Equal provides support for the go-cmp package and testing.
func (e *Error) Equal(e2 *Error) bool {
	return e.Code == e2.Code && e.Message == e2.Message
}

// =============================================================================

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Err   string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// NewFieldErrors creates a field errors.
func NewFieldErrors(field string, err error) *Error {
	fe := FieldErrors{
		{
			Field: field,
			Err:   err.Error(),
		},
	}

	return fe.ToError()
}

// Add adds a field error to the collection.
func (fe *FieldErrors) Add(field string, err error) {
	*fe = append(*fe, FieldError{
		Field: field,
		Err:   err.Error(),
	})
}

// ToError converts the field errors to an Error.
func (fe FieldErrors) ToError() *Error {
	return New(InvalidArgument, fe)
}

// Error implements the error interface.
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}
