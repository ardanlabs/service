package conf

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ErrInvalidStruct indicates that a configuration struct is not the correct type.
var ErrInvalidStruct = errors.New("configuration must be a struct pointer")

// A FieldError occurs when an error occurs updating an individual field
// in the provided struct value.
type FieldError struct {
	fieldName string
	typeName  string
	value     string
	err       error
}

func (err *FieldError) Error() string {
	return fmt.Sprintf("conf: error assigning to field %s: converting '%s' to type %s. details: %s", err.fieldName, err.value, err.typeName, err.err)
}

// Sourcer provides the ability to source data from a configuration source.
// Consider the use of lazy-loading for sourcing large datasets or systems.
type Sourcer interface {

	// Source takes the field key and attempts to locate that key in its
	// configuration data. Returns true if found with the value.
	Source(fld field) (string, bool)
}

// Parse parses configuration into the provided struct.
func Parse(args []string, namespace string, cfgStruct interface{}, sources ...Sourcer) error {

	// Create the flag source.
	flag, err := newSourceFlag(args)
	if err != nil {
		return err
	}

	// Append default sources to any provided list.
	sources = append(sources, newSourceEnv(namespace))
	sources = append(sources, flag)

	// Get the list of fields from the configuration struct to process.
	fields, err := extractFields(nil, cfgStruct)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return errors.New("no fields identified in config struct")
	}

	// Process all fields found in the config struct provided.
	for _, field := range fields {

		// If the field is supposed to hold the leftover args then copy them in
		// from the flags source.
		if field.field.Type() == argsT {
			args := reflect.ValueOf(Args(flag.args))
			field.field.Set(args)
			continue
		}

		// Set any default value into the struct for this field.
		if field.options.defaultVal != "" {
			if err := processField(field.options.defaultVal, field.field); err != nil {
				return &FieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     field.options.defaultVal,
					err:       err,
				}
			}
		}

		// Process each field against all sources.
		var provided bool
		for _, sourcer := range sources {
			if sourcer == nil {
				continue
			}

			var value string
			if value, provided = sourcer.Source(field); !provided {
				continue
			}

			// A value was found so update the struct value with it.
			if err := processField(value, field.field); err != nil {
				return &FieldError{
					fieldName: field.name,
					typeName:  field.field.Type().String(),
					value:     value,
					err:       err,
				}
			}
		}

		// If this key is not provided by any source, check if it was
		// required to be provided.
		if !provided && field.options.required {
			return fmt.Errorf("required field %s is missing value", field.name)
		}
	}

	return nil
}

// Usage provides output to display the config usage on the command line.
func Usage(namespace string, v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}

	return fmtUsage(namespace, fields), nil
}

// String returns a stringified version of the provided conf-tagged
// struct, minus any fields tagged with `noprint`.
func String(v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}

	var s strings.Builder
	for i, fld := range fields {
		if !fld.options.noprint {
			s.WriteString(flagUsage(fld))
			s.WriteString("=")
			s.WriteString(fmt.Sprintf("%v", fld.field.Interface()))
			if i < len(fields)-1 {
				s.WriteString("\n")
			}
		}
	}

	return s.String(), nil
}

// Args holds command line arguments after flags have been parsed.
type Args []string

// argsT is used by Parse and Usage to detect struct fields of the Args type.
var argsT = reflect.TypeOf(Args{})

// Num returns the i'th argument in the Args slice. It returns an empty string
// the request element is not present.
func (a Args) Num(i int) string {
	if i < 0 || i >= len(a) {
		return ""
	}
	return a[i]
}
