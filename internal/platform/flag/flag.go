package flag

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ErrHelp is provided to identify when help is being displayed.
var ErrHelp = errors.New("providing help")

// Process compares the specified command line arguments against the provided
// struct value and updates the fields that are identified.
func Process(v interface{}) error {
	if len(os.Args) == 1 {
		return nil
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Print(display(os.Args[0], v))
		return ErrHelp
	}

	args, err := parse("", v)
	if err != nil {
		return err
	}

	if err := apply(os.Args, args); err != nil {
		return err
	}

	return nil
}

// display provides a pretty print display of the command line arguments.
func display(appName string, v interface{}) string {
	/*
		Current display format for a field.
		Usage of <app name>
		-short --long type	<default> : description
		-a --web_apihost string  <0.0.0.0:3000> : The ip:port for the api endpoint.
	*/

	args, err := parse("", v)
	if err != nil {
		return fmt.Sprint("unable to display help", err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nUsage of %s\n", appName))
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

// configArg represents a single argument for a given field
// in the config structure.
type configArg struct {
	Short   string
	Long    string
	Default string
	Type    string
	Desc    string

	field reflect.Value
}

// parse will reflect over the provided struct value and build a
// collection of all possible config arguments.
func parse(parentField string, v interface{}) ([]configArg, error) {

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

	var cfgArgs []configArg

	// We need to iterate over the fields of the struct value we are processing.
	// If the field is a struct then recurse to process its fields. If we have
	// a field that is not a struct, pull the metadata. The `field` field is
	// important because it is how we update things later.
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.Type.Kind() == reflect.Struct {
			args, err := parse(parentField+field.Name, val.Field(i))
			if err != nil {
				return nil, err
			}
			cfgArgs = append(cfgArgs, args...)
			continue
		}

		cfgArg := configArg{
			Short:   field.Tag.Get("flag"),
			Long:    parentField + strings.ToLower(field.Name),
			Type:    field.Type.Name(),
			Default: field.Tag.Get("default"),
			Desc:    field.Tag.Get("flagdesc"),
			field:   val.Field(i),
		}
		cfgArgs = append(cfgArgs, cfgArg)
	}

	return cfgArgs, nil
}

// apply reads the command line arguments and applies any overrides to
// the provided struct value.
func apply(osArgs []string, cfgArgs []configArg) (err error) {

	// There is so much room for panics here it hurts.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unhandled exception %v", r)
		}
	}()

	lArgs := len(osArgs[1:])
	for i := 1; i <= lArgs; i++ {
		osArg := osArgs[i]

		// Capture the next flag.
		var flag string
		switch {
		case strings.HasPrefix(osArg, "-test"):
			return nil
		case strings.HasPrefix(osArg, "--"):
			flag = osArg[2:]
		case strings.HasPrefix(osArg, "-"):
			flag = osArg[1:]
		default:
			return fmt.Errorf("invalid command line %q", osArg)
		}

		// Is this flag represented in the config struct.
		var cfgArg configArg
		for _, arg := range cfgArgs {
			if arg.Short == flag || arg.Long == flag {
				cfgArg = arg
				break
			}
		}

		// Did we find this flag represented in the struct?
		if !cfgArg.field.IsValid() {
			return fmt.Errorf("unknown flag %q", flag)
		}

		if cfgArg.Type == "bool" {
			if err := update(cfgArg, ""); err != nil {
				return err
			}
			continue
		}

		// Capture the value for this flag.
		i++
		value := osArgs[i]

		// Process the struct value.
		if err := update(cfgArg, value); err != nil {
			return err
		}
	}

	return nil
}

// update applies the value provided on the command line to the struct.
func update(cfgArg configArg, value string) error {
	switch cfgArg.Type {
	case "string":
		cfgArg.field.SetString(value)
	case "int":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("unable to convert value %q to int", value)
		}
		cfgArg.field.SetInt(int64(i))
	case "Duration":
		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("unable to convert value %q to duration", value)
		}
		cfgArg.field.SetInt(int64(d))
	case "bool":
		cfgArg.field.SetBool(true)
	default:
		return fmt.Errorf("type not supported %q", cfgArg.Type)
	}

	return nil
}
