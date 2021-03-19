package conf

import (
	"errors"
	"fmt"
	"net/url"
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
	Source(fld Field) (string, bool)
}

// Version provides the abitily to add version and description to the application.
type Version struct {
	SVN  string
	Desc string
}

// VersionString provides output to display the application version and description on the command line.
func VersionString(namespace string, v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	for i := range fields {
		if fields[i].Name == versionKey && fields[i].Field.Len() > 0 {
			str.WriteString("Version: ")
			str.WriteString(fields[i].Field.String())
			continue
		}
		if fields[i].Name == descKey && fields[i].Field.Len() > 0 {
			if str.Len() > 0 {
				str.WriteString("\n")
			}
			str.WriteString(fields[i].Field.String())
			break
		}
	}
	return str.String(), nil
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
		if field.Field.Type() == argsT {
			args := reflect.ValueOf(Args(flag.args))
			field.Field.Set(args)
			continue
		}

		// Set any default value into the struct for this field.
		if field.Options.DefaultVal != "" {
			if err := processField(field.Options.DefaultVal, field.Field); err != nil {
				return &FieldError{
					fieldName: field.Name,
					typeName:  field.Field.Type().String(),
					value:     field.Options.DefaultVal,
					err:       err,
				}
			}
		}

		// Process each field against all sources.
		var everProvided bool
		for _, sourcer := range sources {
			if sourcer == nil {
				continue
			}

			value, provided := sourcer.Source(field)
			if !provided {
				continue
			}
			everProvided = true

			// A value was found so update the struct value with it.
			if err := processField(value, field.Field); err != nil {
				return &FieldError{
					fieldName: field.Name,
					typeName:  field.Field.Type().String(),
					value:     value,
					err:       err,
				}
			}
		}

		// If this key is not provided by any source, check if it was
		// required to be provided.
		if !everProvided && field.Options.Required {
			return fmt.Errorf("required field %s is missing value", field.Name)
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
		if fld.Options.Noprint {
			continue
		}

		s.WriteString(flagUsage(fld))
		s.WriteString("=")
		v := fmt.Sprintf("%v", fld.Field.Interface())

		switch {
		case fld.Options.Mask:
			if u, err := url.Parse(v); err == nil {
				userPass := u.User.String()
				if userPass != "" {
					v = strings.Replace(v, userPass, "xxxxxx:xxxxxx", 1)
					s.WriteString(v)
					break
				}
			}
			s.WriteString("xxxxxx")

		default:
			s.WriteString(v)
		}

		if i < len(fields)-1 {
			s.WriteString("\n")
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
