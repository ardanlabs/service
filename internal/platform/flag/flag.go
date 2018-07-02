package flag

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ErrHelp is provided to identify when help is being displayed.
var ErrHelp = errors.New("providing help")

// Process compares the specified command line arguments against the provided
// struct value and updates the fields that are identified.
func Process(osArgs []string, v interface{}) error {
	if osArgs[1] == "-h" || osArgs[1] == "--help" {
		fmt.Print(display(v))
		return ErrHelp
	}

	args, err := parse("", v)
	if err != nil {
		return err
	}

	if err := apply(osArgs, args); err != nil {
		return err
	}

	return nil
}

// display provides a pretty print display of the command line arguments.
func display(v interface{}) string {
	/*
		Current display format for a field.
		-short --long type	<default> : description
		-a --web_apihost string  <0.0.0.0:3000> : The ip:port for the api endpoint.
	*/

	args, err := parse("", v)
	if err != nil {
		return fmt.Sprint("unable to display help", err)
	}

	var b strings.Builder
	b.WriteString("\n")
	for _, arg := range args {
		if arg.Short != "" {
			b.WriteString(fmt.Sprintf("-%s ", arg.Short))
		}
		b.WriteString(fmt.Sprintf("--%s %s", arg.Long, arg.Type))
		if arg.Default != "" {
			b.WriteString(fmt.Sprintf("  <%s>", arg.Default))
		}
		if arg.Desc != "" {
			b.WriteString(fmt.Sprintf(" : %s", arg.Desc))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// argument represents a single argument for a given flag.
type argument struct {
	Short   string
	Long    string
	Default string
	Type    string
	Desc    string

	field reflect.Value
}

// parse will reflect over the provided struct value and build a
// collection of argument metadata for the command line.
func parse(parentField string, v interface{}) ([]argument, error) {

	// Reflect on the value to get started.
	rawValue := reflect.ValueOf(v)

	// If a parent field is provided we are recursing. We are now
	// processing a struct within a struct. We need the parent struct
	// name for namespacing.
	if parentField != "" {
		parentField = strings.ToLower(parentField) + "_"
	}

	// We need to check we have a pointer else we can't modify anything
	// later. With the pointer, get the value that the pointer points to.
	// With a struct, that means we are recursing and we need to assert to
	// get the inner struct value to process it.
	var val reflect.Value
	switch rawValue.Kind() {
	case reflect.Ptr:
		val = rawValue.Elem()
		if val.Kind() != reflect.Struct {
			return nil, fmt.Errorf("incompatible type `%v` looking for a pointer", val.Kind())
		}
	case reflect.Struct:
		var ok bool
		if val, ok = v.(reflect.Value); !ok {
			return nil, fmt.Errorf("internal recurse error")
		}
	default:
		return nil, fmt.Errorf("incompatible type `%v`", rawValue.Kind())
	}

	var args []argument

	// We need to iterate over the fields of the struct value we are processing.
	// If the field is a struct then recurse to process its fields. If we have
	// a field that is not a struct, get pull the metadata. The `field` field
	// is important because it is how we update things later.
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Type.Kind() == reflect.Struct {
			newArgs, err := parse(parentField+field.Name, val.Field(i))
			if err != nil {
				return nil, err
			}
			args = append(args, newArgs...)
			continue
		}

		arg := argument{
			Short:   field.Tag.Get("flag"),
			Long:    parentField + strings.ToLower(field.Name),
			Type:    field.Type.Name(),
			Default: field.Tag.Get("default"),
			Desc:    field.Tag.Get("flagdesc"),
			field:   val.Field(i),
		}
		args = append(args, arg)
	}

	return args, nil
}

// apply reads the command line arguments and applies any overrides to
// the provided struct value.
func apply(osArgs []string, args []argument) (err error) {

	// There is so much room for panics here it hurts.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unhandled exception %v", r)
		}
	}()

	var field string
	var value string

	for _, osArg := range osArgs[1:] {

		// Need to find a field and value combination.
		switch {
		case strings.HasPrefix(osArg, "--"):
			field = osArg[2:]
		case strings.HasPrefix(osArg, "-"):
			field = osArg[1:]
		default:
			value = osArg
		}

		// Process the combination.
		if field != "" && value != "" {
			for _, arg := range args {

				// Update the struct value on a match.
				if arg.Short == field || arg.Long == field {
					switch arg.Type {
					case "string":
						arg.field.SetString(value)
					case "int":
						i, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("unable to convert value %q to int", value)
						}
						arg.field.SetInt(int64(i))
					case "Duration":
						d, err := time.ParseDuration(value)
						if err != nil {
							return fmt.Errorf("unable to convert value %q to duration", value)
						}
						arg.field.SetInt(int64(d))
					}
				}
			}

			// Reset to find the next one.
			field = ""
			value = ""
		}
	}
	return nil
}
