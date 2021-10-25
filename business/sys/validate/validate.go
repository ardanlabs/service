// Package validate contains the support for validating models.
package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/google/uuid"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

// translator is a cache of locale and translation information.
var translator ut.Translator

func init() {

	// Instantiate a validator.
	validate = validator.New()

	// Create a translator for english so the error messages are
	// more human readable than technical.
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")

	// Register the english error messages for use.
	en_translations.RegisterDefaultTranslations(validate, translator)

	// Use JSON tag names for errors instead of Go struct names.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Check validates the provided model against it's declared tags.
func Check(val interface{}) error {
	if err := validate.Struct(val); err != nil {

		// Use errors.As to get the real error value.
		var vErrors validator.ValidationErrors
		if !errors.As(err, &vErrors) {
			return err
		}

		var fields FieldErrors
		for _, verror := range vErrors {
			field := FieldError{
				Field: verror.Field(),
				Error: verror.Translate(translator),
			}
			fields = append(fields, field)
		}

		return fields
	}

	return nil
}

// GenerateID generate a unique id for entities.
func GenerateID() string {
	return uuid.NewString()
}

// CheckID validates that the format of an id is valid.
func CheckID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("ID is not in its proper form")
	}
	return nil
}
