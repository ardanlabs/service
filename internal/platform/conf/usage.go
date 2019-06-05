package conf

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"text/tabwriter"
)

func fmtUsage(namespace string, fields []field) string {
	var sb strings.Builder

	fields = append(fields, field{
		name:      "help",
		boolField: true,
		field:     reflect.ValueOf(true),
		flagKey:   []string{"help"},
		options: fieldOptions{
			shortFlagChar: 'h',
			help:          "display this help message",
		}})

	_, file := path.Split(os.Args[0])
	fmt.Fprintf(&sb, "Usage: %s [options] [arguments]\n\n", file)

	fmt.Fprintln(&sb, "OPTIONS")
	w := new(tabwriter.Writer)
	w.Init(&sb, 0, 4, 2, ' ', tabwriter.TabIndent)

	for _, fld := range fields {

		// Skip printing usage info for fields that just hold arguments.
		if fld.field.Type() == argsT {
			continue
		}

		fmt.Fprintf(w, "  %s", flagUsage(fld))

		// Do not display env vars for help since they aren't respected.
		if fld.name != "help" {
			fmt.Fprintf(w, "/%s", envUsage(namespace, fld))
		}

		typeName, help := getTypeAndHelp(&fld)

		// Do not display type info for help because it would show <bool> but our
		// parsing does not really treat --help as a boolean field. Its presence
		// always indicates true even if they do --help=false.
		if fld.name != "help" {
			fmt.Fprintf(w, "\t%s", typeName)
		}

		fmt.Fprintf(w, "\t%s\n", getOptString(fld))
		if help != "" {
			fmt.Fprintf(w, "  %s\n", help)
		}
	}

	w.Flush()
	return sb.String()
}

// getTypeAndHelp extracts the type and help message for a single field for
// printing in the usage message. If the help message contains text in
// single quotes ('), this is assumed to be a more specific "type", and will
// be returned as such. If there are no back quotes, it attempts to make a
// guess as to the type of the field. Boolean flags are not printed with a
// type, manually-specified or not, since their presence is equated with a
// 'true' value and their absence with a 'false' value. If a type cannot be
// determined, it will simply give the name "value". Slices will be annotated
// as "<Type>,[Type...]", where "Type" is whatever type name was chosen.
// (adapted from package flag).
func getTypeAndHelp(fld *field) (name string, usage string) {

	// Look for a single-quoted name.
	usage = fld.options.help
	for i := 0; i < len(usage); i++ {
		if usage[i] == '\'' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '\'' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
				}
			}
			break // Only one single quote; use type name.
		}
	}

	var isSlice bool
	if fld.field.IsValid() {
		t := fld.field.Type()

		// If it's a pointer, we want to deref.
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// If it's a slice, we want the type of the slice elements.
		if t.Kind() == reflect.Slice {
			t = t.Elem()
			isSlice = true
		}

		// If no explicit name was provided, attempt to get the type
		if name == "" {
			switch t.Kind() {
			case reflect.Bool:
				name = "bool"
			case reflect.Float32, reflect.Float64:
				name = "float"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				typ := fld.field.Type()
				if typ.PkgPath() == "time" && typ.Name() == "Duration" {
					name = "duration"
				} else {
					name = "int"
				}
			case reflect.String:
				name = "string"
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				name = "uint"
			default:
				name = "value"
			}
		}
	}

	switch {
	case isSlice:
		name = fmt.Sprintf("<%s>,[%s...]", name, name)
	case name != "":
		name = fmt.Sprintf("<%s>", name)
	default:
	}
	return
}

func getOptString(fld field) string {
	opts := make([]string, 0, 3)
	if fld.options.required {
		opts = append(opts, "required")
	}
	if fld.options.noprint {
		opts = append(opts, "noprint")
	}
	if fld.options.defaultVal != "" {
		opts = append(opts, fmt.Sprintf("default: %s", fld.options.defaultVal))
	}
	if len(opts) > 0 {
		return fmt.Sprintf("(%s)", strings.Join(opts, `,`))
	}
	return ""
}
