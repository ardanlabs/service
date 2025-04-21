package conf

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
)

const (
	buildKey   = "Build"
	descKey    = "Desc"
	helpKey    = "help"
	versionKey = "version"
)

type sortedFields struct {
	fields []Field
}

func (sf sortedFields) Len() int {
	return len(sf.fields)
}

func (sf sortedFields) Swap(i, j int) {
	sf.fields[i], sf.fields[j] = sf.fields[j], sf.fields[i]
}

func (sf sortedFields) Less(i, j int) bool {
	s1 := strings.ToLower(strings.Join(sf.fields[i].FlagKey, `-`))
	s2 := strings.ToLower(strings.Join(sf.fields[j].FlagKey, `-`))

	return s1 < s2
}

func containsField(fields []Field, name string) bool {
	for i := range fields {
		if name == fields[i].Name {
			return true
		}
	}
	return false
}

func fmtUsage(namespace string, fields []Field) string {
	var sb strings.Builder

	fields = append(fields, Field{
		Name:      "help",
		BoolField: true,
		Field:     reflect.ValueOf(true),
		FlagKey:   []string{"help"},
		Options: FieldOptions{
			ShortFlagChar: 'h',
			Help:          "display this help message",
		}})

	if containsField(fields, buildKey) {
		fields = append(fields, Field{
			Name:      "version",
			BoolField: true,
			Field:     reflect.ValueOf(true),
			FlagKey:   []string{"version"},
			Options: FieldOptions{
				ShortFlagChar: 'v',
				Help:          "display version",
			}})
	}

	sf := sortedFields{fields: fields}
	sort.Sort(&sf)

	_, file := path.Split(os.Args[0])
	fmt.Fprintf(&sb, "Usage: %s [options...] [arguments...]\n\n", file)

	w := new(tabwriter.Writer)
	w.Init(&sb, 0, 4, 2, ' ', tabwriter.TabIndent)

	fmt.Fprintln(&sb, "OPTIONS")
	writeOptions(w, sf.fields)

	fmt.Fprintln(&sb, "ENVIRONMENT")
	writeEnv(w, namespace, sf.fields)

	return sb.String()
}

func writeOptions(w *tabwriter.Writer, fields []Field) {
	for _, fld := range fields {

		// Skip printing usage info for fields that just hold arguments.
		if fld.Field.Type() == argsT {
			continue
		}

		// Do not display version fields SVN and Description
		if fld.Name == buildKey || fld.Name == descKey {
			continue
		}

		fmt.Fprintf(w, "  %s", flagUsage(fld))

		typeName, help := getTypeAndHelp(&fld)

		// Do not display type info for help because it would show <bool> but our
		// parsing does not really treat --help as a boolean field. Its presence
		// always indicates true even if they do --help=false.
		if fld.Name != helpKey && fld.Name != versionKey {
			fmt.Fprintf(w, "\t%s", typeName)
			fmt.Fprintf(w, "\t%s", getOptString(fld))
		} else {
			fmt.Fprint(w, "\t\t")
		}

		fmt.Fprintf(w, "\t%s", help)

		fmt.Fprint(w, "\n")
	}

	fmt.Fprint(w, "\n")
	w.Flush()
}

func writeEnv(w *tabwriter.Writer, namespace string, fields []Field) {
	for _, fld := range fields {

		// Skip printing usage info for fields that just hold arguments.
		if fld.Field.Type() == argsT {
			continue
		}

		// Do not display version fields and Description
		// Do not display env vars for help since they aren't respected.
		if fld.Name == buildKey || fld.Name == descKey ||
			fld.Name == helpKey || fld.Name == versionKey {
			continue
		}

		fmt.Fprintf(w, "  %s", envUsage(namespace, fld))

		typeName, help := getTypeAndHelp(&fld)

		fmt.Fprintf(w, "\t%s", typeName)
		fmt.Fprintf(w, "\t%s", getOptString(fld))

		fmt.Fprintf(w, "\t%s", help)

		fmt.Fprint(w, "\n")
	}

	w.Flush()
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
func getTypeAndHelp(fld *Field) (name string, usage string) {

	// Look for a single-quoted name.
	usage = fld.Options.Help
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
	if fld.Field.IsValid() {
		t := fld.Field.Type()

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
				typ := fld.Field.Type()
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

func getOptString(fld Field) string {
	opts := make([]string, 0, 3)
	if fld.Options.Required {
		opts = append(opts, "required")
	}
	if fld.Options.NotZero {
		opts = append(opts, "notzero")
	}
	if fld.Options.Noprint {
		opts = append(opts, "noprint")
	}
	if fld.Options.Immutable {
		opts = append(opts, "immutable")
	}
	if fld.Options.Mask {
		fld.Options.DefaultVal = maskVal(fld.Options.DefaultVal)
	}
	if fld.Options.DefaultVal != "" {
		opts = append(opts, fmt.Sprintf("default: %s", fld.Options.DefaultVal))
	}
	if len(opts) > 0 {
		return fmt.Sprintf("(%s)", strings.Join(opts, `,`))
	}
	return ""
}
