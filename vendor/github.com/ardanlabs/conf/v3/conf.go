package conf

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
)

// ErrInvalidStruct indicates that a configuration struct is not the correct type.
var ErrInvalidStruct = errors.New("configuration must be a struct pointer")

// Version provides the abitily to add version and description to the application.
type Version struct {
	Build string
	Desc  string
}

// Parsers declare behavior to extend the different parsers that
// can be used to unmarshal config.
type Parsers interface {
	Process(prefix string, cfg interface{}) error
}

// =============================================================================

// Parse parses the specified config struct. This function will
// apply the defaults first and then apply environment variables and
// command line argument overrides to the struct. ErrHelpWanted is
// returned when the --help or --version are detected.
func Parse(prefix string, cfg interface{}, parsers ...Parsers) (string, error) {
	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	for _, parser := range parsers {
		if err := parser.Process(prefix, cfg); err != nil {
			return "", fmt.Errorf("external parser: %w", err)
		}
	}

	err := parse(args, prefix, cfg)
	if err == nil {
		return "", nil
	}

	switch err {
	case ErrHelpWanted:
		usage, err := UsageInfo(prefix, cfg)
		if err != nil {
			return "", fmt.Errorf("generating config usage: %w", err)
		}
		return usage, ErrHelpWanted

	case errVersionWanted:
		version, err := VersionInfo(prefix, cfg)
		if err != nil {
			return "", fmt.Errorf("generating config version: %w", err)
		}

		return version, ErrHelpWanted
	}

	return "", fmt.Errorf("parsing config: %w", err)
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
			s.WriteString(maskVal(v))

		default:
			s.WriteString(v)
		}

		if i < len(fields)-1 {
			s.WriteString("\n")
		}
	}

	return s.String(), nil
}

// maskVal masks an entire string or the user:password pair of a URL.
func maskVal(v string) string {
	mask := "xxxxxx"
	if u, err := url.Parse(v); err == nil {
		userPass := u.User.String()
		if userPass != "" {
			mask = strings.Replace(v, userPass, "xxxxxx:xxxxxx", 1)
		}
	}
	return mask
}

// UsageInfo provides output to display the config usage on the command line.
func UsageInfo(namespace string, v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}

	return fmtUsage(namespace, fields), nil
}

// VersionInfo provides output to display the application version and description on the command line.
func VersionInfo(namespace string, v interface{}) (string, error) {
	fields, err := extractFields(nil, v)
	if err != nil {
		return "", err
	}

	var str strings.Builder
	for i := range fields {
		if fields[i].Name == buildKey && fields[i].Field.Len() > 0 {
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

// =============================================================================

// parse parses configuration into the provided struct.
func parse(args []string, namespace string, cfgStruct interface{}) error {
	// Create the flag and env sources.
	flag, err := newSourceFlag(args)
	if err != nil {
		return err
	}
	sources := []sourcer{newSourceEnv(namespace), flag}

	// Get the list of fields from the configuration struct to process.
	fields, err := extractFields(nil, cfgStruct)
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return errors.New("no fields identified in config struct")
	}

	// Hold the field the is supposed to hold the leftover args.
	var argsF *Field

	// Process all fields found in the config struct provided.
	for _, field := range fields {
		field := field

		// If the field is supposed to hold the leftover args then hold a reference for later.
		if field.Field.Type() == argsT {
			argsF = &field
			continue
		}

		// Set any default value into the struct for this field.
		if field.Options.DefaultVal != "" {
			if err := processField(true, field.Options.DefaultVal, field.Field); err != nil {
				return &FieldError{
					fieldName: field.Name,
					typeName:  field.Field.Type().String(),
					value:     field.Options.DefaultVal,
					err:       err,
				}
			}
		}

		// Flag to check if any value is provided.
		provided := false

		// Process each field against all sources.
		for _, sourcer := range sources {
			if sourcer == nil {
				continue
			}

			value, ok := sourcer.Source(field)
			if !ok {
				continue
			}

			// A value was found so update the struct value with it.
			if err := processField(false, value, field.Field); err != nil {
				return &FieldError{
					fieldName: field.Name,
					typeName:  field.Field.Type().String(),
					value:     value,
					err:       err,
				}
			}

			provided = true
		}

		// If the field is marked 'required', check if no value was provided.
		if field.Options.Required && !provided {
			return fmt.Errorf("required field %s is missing value", field.Name)
		}
	}

	// If there is a field that is supposed to hold the leftover args then copy them in
	// from the flags source.
	if argsF != nil {
		args := reflect.ValueOf(Args(flag.args))
		argsF.Field.Set(args)
	}

	return nil
}

// =============================================================================

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
