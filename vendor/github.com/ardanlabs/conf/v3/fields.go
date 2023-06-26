package conf

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

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

// =============================================================================

// Field maintains information about a field in the configuration struct.
type Field struct {
	Name    string
	FlagKey []string
	EnvKey  []string
	Field   reflect.Value
	Options FieldOptions

	// Important for flag parsing or any other source where
	// booleans might be treated specially.
	BoolField bool
}

// FieldOptions maintain flag options for a given field.
type FieldOptions struct {
	Help          string
	DefaultVal    string
	EnvName       string
	FlagName      string
	ShortFlagChar rune
	Noprint       bool
	Required      bool
	Mask          bool
}

// extractFields uses reflection to examine the struct and generate the keys.
func extractFields(prefix []string, target interface{}) ([]Field, error) {
	if prefix == nil {
		prefix = []string{}
	}
	s := reflect.ValueOf(target)

	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidStruct
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidStruct
	}
	targetType := s.Type()

	var fields []Field

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		structField := targetType.Field(i)

		// Get the conf tags associated with this item (if any).
		fieldTags := structField.Tag.Get("conf")

		// If it's ignored or can't be set, move on.
		if !f.CanSet() || fieldTags == "-" {
			continue
		}

		fieldName := structField.Name

		// Get and options.  TODO: Need more.
		fieldOpts, err := parseTag(fieldTags)
		if err != nil {
			return nil, fmt.Errorf("conf: error parsing tags for field %s: %s", fieldName, err)
		}

		// Generate the field key. This could be ignored.
		fieldKey := append(prefix, camelSplit(fieldName)...)

		// Drill down through pointers until we bottom out at type or nil.
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {

				// It's not a struct so leave it alone.
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}

				// It is a struct so zero it out.
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		switch {

		// If we found a struct that can't deserialize itself, drill down,
		// appending fields as we go.
		case f.Kind() == reflect.Struct && setterFrom(f) == nil && textUnmarshaler(f) == nil && binaryUnmarshaler(f) == nil:

			// Prefix for any subkeys is the fieldKey, unless it's
			// anonymous, then it's just the prefix so far.
			innerPrefix := fieldKey
			if structField.Anonymous {
				innerPrefix = prefix
			}

			embeddedPtr := f.Addr().Interface()
			innerFields, err := extractFields(innerPrefix, embeddedPtr)
			if err != nil {
				return nil, err
			}
			fields = append(fields, innerFields...)

		default:
			envKey := make([]string, len(fieldKey))
			copy(envKey, fieldKey)
			if fieldOpts.EnvName != "" {
				envKey = strings.Split(fieldOpts.EnvName, "_")
			}

			flagKey := make([]string, len(fieldKey))
			copy(flagKey, fieldKey)
			if fieldOpts.FlagName != "" {
				flagKey = strings.Split(fieldOpts.FlagName, "-")
			}

			fld := Field{
				Name:      fieldName,
				EnvKey:    envKey,
				FlagKey:   flagKey,
				Field:     f,
				Options:   fieldOpts,
				BoolField: f.Kind() == reflect.Bool,
			}
			fields = append(fields, fld)
		}
	}

	return fields, nil
}

func parseTag(tagStr string) (FieldOptions, error) {
	var f FieldOptions
	if tagStr == "" {
		return f, nil
	}

	tagParts := strings.Split(tagStr, ",")
	for _, tagPart := range tagParts {
		vals := strings.SplitN(tagPart, ":", 2)
		tagProp := vals[0]

		switch len(vals) {
		case 1:
			switch tagProp {
			case "noprint":
				f.Noprint = true
			case "required":
				f.Required = true
			case "mask":
				f.Mask = true
			}
		case 2:
			tagPropVal := strings.TrimSpace(vals[1])
			if tagPropVal == "" {
				return f, fmt.Errorf("tag %q missing a value", tagProp)
			}
			switch tagProp {
			case "short":
				if len([]rune(tagPropVal)) != 1 {
					return f, fmt.Errorf("short value must be a single rune, got %q", tagPropVal)
				}
				f.ShortFlagChar = []rune(tagPropVal)[0]
			case "default":
				f.DefaultVal = tagPropVal
			case "env":
				f.EnvName = tagPropVal
			case "flag":
				f.FlagName = tagPropVal
			case "help":
				f.Help = tagPropVal
			}
		default:
			// TODO: Do we check for integrity issues here?
		}
	}

	// Perform a sanity check.
	switch {
	case f.Required && f.DefaultVal != "":
		return f, fmt.Errorf("cannot set both `required` and `default`")
	}

	return f, nil
}

// camelSplit takes a string based on camel case and splits it.
func camelSplit(src string) []string {
	if src == "" {
		return []string{}
	}
	if len(src) < 2 {
		return []string{src}
	}

	runes := []rune(src)

	lastClass := charClass(runes[0])
	lastIdx := 0
	out := []string{}

	// Split into fields based on class of unicode character.
	for i, r := range runes {
		class := charClass(r)

		// If the class has transitioned.
		if class != lastClass {

			// If going from uppercase to lowercase, we want to retain the last
			// uppercase letter for names like FOOBar, which should split to
			// FOO Bar.
			switch {
			case lastClass == classUpper && class != classNumber:
				if i-lastIdx > 1 {
					out = append(out, string(runes[lastIdx:i-1]))
					lastIdx = i - 1
				}
			default:
				out = append(out, string(runes[lastIdx:i]))
				lastIdx = i
			}
		}

		if i == len(runes)-1 {
			out = append(out, string(runes[lastIdx:]))
		}
		lastClass = class
	}

	return out
}

func processField(settingDefault bool, value string, field reflect.Value) error {
	typ := field.Type()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}

	// We don't want a default value to override a
	// proper setting.
	if settingDefault && !field.IsZero() {
		return nil
	}

	// Look for a Set method.
	setter := setterFrom(field)
	if setter != nil {
		return setter.Set(value)
	}

	if t := textUnmarshaler(field); t != nil {
		return t.UnmarshalText([]byte(value))
	}

	if b := binaryUnmarshaler(field); b != nil {
		return b.UnmarshalBinary([]byte(value))
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var (
			val int64
			err error
		)
		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			var d time.Duration
			d, err = time.ParseDuration(value)
			val = int64(d)
		} else {
			val, err = strconv.ParseInt(value, 0, typ.Bits())
		}
		if err != nil {
			return err
		}

		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Slice:
		vals := strings.Split(value, ";")
		sl := reflect.MakeSlice(typ, len(vals), len(vals))
		for i, val := range vals {
			err := processField(false, val, sl.Index(i))
			if err != nil {
				return err
			}
		}
		field.Set(sl)
	case reflect.Map:
		mp := reflect.MakeMap(typ)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, ";")
			for _, pair := range pairs {
				kvpair := strings.Split(pair, ":")
				if len(kvpair) != 2 {
					return fmt.Errorf("invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				err := processField(false, kvpair[0], k)
				if err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				err = processField(false, kvpair[1], v)
				if err != nil {
					return err
				}
				mp.SetMapIndex(k, v)
			}
		}
		field.Set(mp)
	}
	return nil
}

func interfaceFrom(field reflect.Value, fn func(interface{}, *bool)) {

	// It may be impossible for a struct field to fail this check.
	if !field.CanInterface() {
		return
	}

	var ok bool
	fn(field.Interface(), &ok)
	if !ok && field.CanAddr() {
		fn(field.Addr().Interface(), &ok)
	}
}

// Setter is implemented by types can self-deserialize values.
// Any type that implements flag.Value also implements Setter.
type Setter interface {
	Set(value string) error
}

func setterFrom(field reflect.Value) (s Setter) {
	interfaceFrom(field, func(v interface{}, ok *bool) { s, *ok = v.(Setter) })
	return s
}

func textUnmarshaler(field reflect.Value) (t encoding.TextUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { t, *ok = v.(encoding.TextUnmarshaler) })
	return t
}

func binaryUnmarshaler(field reflect.Value) (b encoding.BinaryUnmarshaler) {
	interfaceFrom(field, func(v interface{}, ok *bool) { b, *ok = v.(encoding.BinaryUnmarshaler) })
	return b
}

const (
	classLower int = iota
	classUpper
	classNumber
	classOther
)

func charClass(r rune) int {
	switch {
	case unicode.IsLower(r):
		return classLower
	case unicode.IsUpper(r):
		return classUpper
	case unicode.IsDigit(r):
		return classNumber
	}
	return classOther
}
