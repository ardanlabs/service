package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	validator "gopkg.in/go-playground/validator.v8"
)

// validate provides a validator for checking models.
var validate = validator.New(&validator.Config{
	TagName:      "validate",
	FieldNameTag: "json",
})

// Invalid describes a validation error belonging to a specific field.
type Invalid struct {
	Fld string `json:"field_name"`
	Err string `json:"error"`
}

// InvalidError is a custom error type for invalid fields.
type InvalidError []Invalid

// Error implements the error interface for InvalidError.
func (err InvalidError) Error() string {
	var str string
	for _, v := range err {
		str = fmt.Sprintf("%s,{%s:%s}", str, v.Fld, v.Err)
	}
	return str
}

// Unmarshal decodes the input to the struct type and checks the
// fields to verify the value is in a proper state.
func Unmarshal(r io.Reader, v interface{}) error {
	decode := json.NewDecoder(r)
	decode.DisallowUnknownFields()
	if err := decode.Decode(v); err != nil {
		return ErrorWithStatus(err, http.StatusBadRequest)
	}

	var inv InvalidError
	if fve := validate.Struct(v); fve != nil {
		for _, fe := range fve.(validator.ValidationErrors) {
			inv = append(inv, Invalid{Fld: fe.Field, Err: fe.Tag})
		}
		return ErrorWithStatus(inv, http.StatusBadRequest)
	}

	return nil
}
