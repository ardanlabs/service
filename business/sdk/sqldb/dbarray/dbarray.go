/*
Code taken from https://github.com/lib/pq

Copyright (c) 2011-2013, 'pq' Contributors Portions Copyright (C) 2011 Blake Mizerany

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the
Software, and to permit persons to whom the Software is furnished to do so, subject
to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

// Package dbarray provides support for database array types.
package dbarray

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var typeByteSlice = reflect.TypeOf([]byte{})
var typeDriverValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
var typeSQLScanner = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

// Array returns the optimal driver.Valuer and sql.Scanner for an array or
// slice of any dimension.
//
// For example:
//
//	db.Query(`SELECT * FROM t WHERE id = ANY($1)`, pq.Array([]int{235, 401}))
//
//	var x []sql.NullInt64
//	db.QueryRow(`SELECT ARRAY[235, 401]`).Scan(pq.Array(&x))
//
// Scanning multi-dimensional arrays is not supported.  Arrays where the lower
// bound is not one (such as `[0:0]={1}') are not supported.
func Array(a any) interface {
	driver.Valuer
	sql.Scanner
} {
	switch a := a.(type) {
	case []bool:
		return (*Bool)(&a)
	case []float64:
		return (*Float64)(&a)
	case []float32:
		return (*Float32)(&a)
	case []int64:
		return (*Int64)(&a)
	case []int32:
		return (*Int32)(&a)
	case []string:
		return (*String)(&a)
	case [][]byte:
		return (*Bytea)(&a)

	case *[]bool:
		return (*Bool)(a)
	case *[]float64:
		return (*Float64)(a)
	case *[]float32:
		return (*Float32)(a)
	case *[]int64:
		return (*Int64)(a)
	case *[]int32:
		return (*Int32)(a)
	case *[]string:
		return (*String)(a)
	case *[][]byte:
		return (*Bytea)(a)
	}

	return Generic{a}
}

// Delimiter may be optionally implemented by driver.Valuer or sql.Scanner
// to override the array delimiter used by Generic.
type Delimiter interface {
	// Delimiter returns the delimiter character(s) for this element's type.
	Delimiter() string
}

// Bool represents a one-dimensional array of the PostgreSQL boolean type.
type Bool []bool

// Scan implements the sql.Scanner interface.
func (a *Bool) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Bool", src)
}

func (a *Bool) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Bool")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Bool, len(elems))
		for i, v := range elems {
			if len(v) != 1 {
				return fmt.Errorf("database: could not parse boolean array index %d: invalid boolean %q", i, v)
			}
			switch v[0] {
			case 't':
				b[i] = true
			case 'f':
				b[i] = false
			default:
				return fmt.Errorf("database: could not parse boolean array index %d: invalid boolean %q", i, v)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Bool) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be exactly two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1+2*n)

		for i := 0; i < n; i++ {
			b[2*i] = ','
			if a[i] {
				b[1+2*i] = 't'
			} else {
				b[1+2*i] = 'f'
			}
		}

		b[0] = '{'
		b[2*n] = '}'

		return string(b), nil
	}

	return "{}", nil
}

// Bytea represents a one-dimensional array of the PostgreSQL bytea type.
type Bytea [][]byte

// Scan implements the sql.Scanner interface.
func (a *Bytea) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Bytea", src)
}

func (a *Bytea) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Bytea")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Bytea, len(elems))
		for i, v := range elems {
			b[i], err = parseBytea(v)
			if err != nil {
				return fmt.Errorf("could not parse bytea array index %d: %s", i, err.Error())
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface. It uses the "hex" format which
// is only supported on PostgreSQL 9.0 or newer.
func (a Bytea) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, 2*N bytes of quotes,
		// 3*N bytes of hex formatting, and N-1 bytes of delimiters.
		size := 1 + 6*n
		for _, x := range a {
			size += hex.EncodedLen(len(x))
		}

		b := make([]byte, size)

		for i, s := 0, b; i < n; i++ {
			o := copy(s, `,"\\x`)
			o += hex.Encode(s[o:], a[i])
			s[o] = '"'
			s = s[o+1:]
		}

		b[0] = '{'
		b[size-1] = '}'

		return string(b), nil
	}

	return "{}", nil
}

// Float64 represents a one-dimensional array of the PostgreSQL double
// precision type.
type Float64 []float64

// Scan implements the sql.Scanner interface.
func (a *Float64) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Float64", src)
}

func (a *Float64) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Float64")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Float64, len(elems))
		for i, v := range elems {
			if b[i], err = strconv.ParseFloat(string(v), 64); err != nil {
				return fmt.Errorf("database: parsing array element index %d: %v", i, err)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Float64) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendFloat(b, a[0], 'f', -1, 64)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendFloat(b, a[i], 'f', -1, 64)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Float32 represents a one-dimensional array of the PostgreSQL double
// precision type.
type Float32 []float32

// Scan implements the sql.Scanner interface.
func (a *Float32) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Float32", src)
}

func (a *Float32) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Float32")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Float32, len(elems))
		for i, v := range elems {
			var x float64
			if x, err = strconv.ParseFloat(string(v), 32); err != nil {
				return fmt.Errorf("database: parsing array element index %d: %v", i, err)
			}
			b[i] = float32(x)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Float32) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendFloat(b, float64(a[0]), 'f', -1, 32)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendFloat(b, float64(a[i]), 'f', -1, 32)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Generic implements the driver.Valuer and sql.Scanner interfaces for
// an array or slice of any dimension.
type Generic struct{ A any }

func (Generic) evaluateDestination(rt reflect.Type) (reflect.Type, func([]byte, reflect.Value) error, string) {
	var assign func([]byte, reflect.Value) error
	var del = ","

	// TODO calculate the assign function for other types
	// TODO repeat this section on the element type of arrays or slices (multidimensional)
	{
		if reflect.PointerTo(rt).Implements(typeSQLScanner) {
			// dest is always addressable because it is an element of a slice.
			assign = func(src []byte, dest reflect.Value) (err error) {
				ss := dest.Addr().Interface().(sql.Scanner)
				if src == nil {
					err = ss.Scan(nil)
				} else {
					err = ss.Scan(src)
				}
				return
			}
			goto FoundType
		}

		assign = func([]byte, reflect.Value) error {
			return fmt.Errorf("database: scanning to %s is not implemented; only sql.Scanner", rt)
		}
	}

FoundType:

	if ad, ok := reflect.Zero(rt).Interface().(Delimiter); ok {
		del = ad.Delimiter()
	}

	return rt, assign, del
}

// Scan implements the sql.Scanner interface.
func (a Generic) Scan(src any) error {
	dpv := reflect.ValueOf(a.A)
	switch {
	case dpv.Kind() != reflect.Ptr:
		return fmt.Errorf("database: destination %T is not a pointer to array or slice", a.A)
	case dpv.IsNil():
		return fmt.Errorf("database: destination %T is nil", a.A)
	}

	dv := dpv.Elem()
	switch dv.Kind() {
	case reflect.Slice:
	case reflect.Array:
	default:
		return fmt.Errorf("database: destination %T is not a pointer to array or slice", a.A)
	}

	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src, dv)
	case string:
		return a.scanBytes([]byte(src), dv)
	case nil:
		if dv.Kind() == reflect.Slice {
			dv.Set(reflect.Zero(dv.Type()))
			return nil
		}
	}

	return fmt.Errorf("database: cannot convert %T to %s", src, dv.Type())
}

func (a Generic) scanBytes(src []byte, dv reflect.Value) error {
	dtype, assign, del := a.evaluateDestination(dv.Type().Elem())
	dims, elems, err := parseArray(src, []byte(del))
	if err != nil {
		return err
	}

	// TODO allow multidimensional

	if len(dims) > 1 {
		return fmt.Errorf("database: scanning from multidimensional ARRAY%s is not implemented",
			strings.Replace(fmt.Sprint(dims), " ", "][", -1))
	}

	// Treat a zero-dimensional array like an array with a single dimension of zero.
	if len(dims) == 0 {
		dims = append(dims, 0)
	}

	for i, rt := 0, dv.Type(); i < len(dims); i, rt = i+1, rt.Elem() {
		switch rt.Kind() {
		case reflect.Slice:
		case reflect.Array:
			if rt.Len() != dims[i] {
				return fmt.Errorf("database: cannot convert ARRAY%s to %s",
					strings.Replace(fmt.Sprint(dims), " ", "][", -1), dv.Type())
			}
		default:
			// TODO handle multidimensional
		}
	}

	values := reflect.MakeSlice(reflect.SliceOf(dtype), len(elems), len(elems))
	for i, e := range elems {
		if err := assign(e, values.Index(i)); err != nil {
			return fmt.Errorf("database: parsing array element index %d: %v", i, err)
		}
	}

	// TODO handle multidimensional

	switch dv.Kind() {
	case reflect.Slice:
		dv.Set(values.Slice(0, dims[0]))
	case reflect.Array:
		for i := 0; i < dims[0]; i++ {
			dv.Index(i).Set(values.Index(i))
		}
	}

	return nil
}

// Value implements the driver.Valuer interface.
func (a Generic) Value() (driver.Value, error) {
	if a.A == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(a.A)

	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return nil, nil
		}
	case reflect.Array:
	default:
		return nil, fmt.Errorf("database: Unable to convert %T to array", a.A)
	}

	if n := rv.Len(); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 0, 1+2*n)

		b, _, err := appendArray(b, rv, n)
		return string(b), err
	}

	return "{}", nil
}

// Int64 represents a one-dimensional array of the PostgreSQL integer types.
type Int64 []int64

// Scan implements the sql.Scanner interface.
func (a *Int64) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Int64", src)
}

func (a *Int64) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Int64")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Int64, len(elems))
		for i, v := range elems {
			if b[i], err = strconv.ParseInt(string(v), 10, 64); err != nil {
				return fmt.Errorf("database: parsing array element index %d: %v", i, err)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Int64) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, a[0], 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, a[i], 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// Int32 represents a one-dimensional array of the PostgreSQL integer types.
type Int32 []int32

// Scan implements the sql.Scanner interface.
func (a *Int32) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to Int32", src)
}

func (a *Int32) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Int32")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Int32, len(elems))
		for i, v := range elems {
			x, err := strconv.ParseInt(string(v), 10, 32)
			if err != nil {
				return fmt.Errorf("database: parsing array element index %d: %v", i, err)
			}
			b[i] = int32(x)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Int32) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendInt(b, int64(a[0]), 10)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendInt(b, int64(a[i]), 10)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// String represents a one-dimensional array of the PostgreSQL character types.
type String []string

// Scan implements the sql.Scanner interface.
func (a *String) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("database: cannot convert %T to String", src)
}

func (a *String) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "String")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(String, len(elems))
		for i, v := range elems {
			if b[i] = string(v); v == nil {
				return fmt.Errorf("database: parsing array element index %d: cannot convert nil to string", i)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a String) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, 2*N bytes of quotes,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+3*n)
		b[0] = '{'

		b = appendArrayQuotedBytes(b, []byte(a[0]))
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = appendArrayQuotedBytes(b, []byte(a[i]))
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// appendArray appends rv to the buffer, returning the extended buffer and
// the delimiter used between elements.
//
// It panics when n <= 0 or rv's Kind is not reflect.Array nor reflect.Slice.
func appendArray(b []byte, rv reflect.Value, n int) ([]byte, string, error) {
	var del string
	var err error

	b = append(b, '{')

	if b, del, err = appendArrayElement(b, rv.Index(0)); err != nil {
		return b, del, err
	}

	for i := 1; i < n; i++ {
		b = append(b, del...)
		if b, del, err = appendArrayElement(b, rv.Index(i)); err != nil {
			return b, del, err
		}
	}

	return append(b, '}'), del, nil
}

// appendArrayElement appends rv to the buffer, returning the extended buffer
// and the delimiter to use before the next element.
//
// When rv's Kind is neither reflect.Array nor reflect.Slice, it is converted
// using driver.DefaultParameterConverter and the resulting []byte or string
// is double-quoted.
//
// See http://www.postgresql.org/docs/current/static/arrays.html#ARRAYS-IO
func appendArrayElement(b []byte, rv reflect.Value) ([]byte, string, error) {
	if k := rv.Kind(); k == reflect.Array || k == reflect.Slice {
		if t := rv.Type(); t != typeByteSlice && !t.Implements(typeDriverValuer) {
			if n := rv.Len(); n > 0 {
				return appendArray(b, rv, n)
			}

			return b, "", nil
		}
	}

	var del = ","
	var err error
	var iv = rv.Interface()

	if ad, ok := iv.(Delimiter); ok {
		del = ad.Delimiter()
	}

	if iv, err = driver.DefaultParameterConverter.ConvertValue(iv); err != nil {
		return b, del, err
	}

	switch v := iv.(type) {
	case nil:
		return append(b, "NULL"...), del, nil
	case []byte:
		return appendArrayQuotedBytes(b, v), del, nil
	case string:
		return appendArrayQuotedBytes(b, []byte(v)), del, nil
	}

	b, err = appendValue(b, iv)
	return b, del, err
}

func appendArrayQuotedBytes(b, v []byte) []byte {
	b = append(b, '"')
	for {
		i := bytes.IndexAny(v, `"\`)
		if i < 0 {
			b = append(b, v...)
			break
		}
		if i > 0 {
			b = append(b, v[:i]...)
		}
		b = append(b, '\\', v[i])
		v = v[i+1:]
	}
	return append(b, '"')
}

func appendValue(b []byte, v driver.Value) ([]byte, error) {
	return append(b, encode(nil, v, 0)...), nil
}

// parseArray extracts the dimensions and elements of an array represented in
// text format. Only representations emitted by the backend are supported.
// Notably, whitespace around brackets and delimiters is significant, and NULL
// is case-sensitive.
//
// See http://www.postgresql.org/docs/current/static/arrays.html#ARRAYS-IO
func parseArray(src, del []byte) (dims []int, elems [][]byte, err error) {
	var depth, i int

	if len(src) < 1 || src[0] != '{' {
		return nil, nil, fmt.Errorf("database: unable to parse array; expected %q at offset %d", '{', 0)
	}

Open:
	for i < len(src) {
		switch src[i] {
		case '{':
			depth++
			i++
		case '}':
			elems = make([][]byte, 0)
			goto Close
		default:
			break Open
		}
	}
	dims = make([]int, i)

Element:
	for i < len(src) {
		switch src[i] {
		case '{':
			if depth == len(dims) {
				break Element
			}
			depth++
			dims[depth-1] = 0
			i++
		case '"':
			var elem = []byte{}
			var escape bool
			for i++; i < len(src); i++ {
				if escape {
					elem = append(elem, src[i])
					escape = false
				} else {
					switch src[i] {
					default:
						elem = append(elem, src[i])
					case '\\':
						escape = true
					case '"':
						elems = append(elems, elem)
						i++
						break Element
					}
				}
			}
		default:
			for start := i; i < len(src); i++ {
				if bytes.HasPrefix(src[i:], del) || src[i] == '}' {
					elem := src[start:i]
					if len(elem) == 0 {
						return nil, nil, fmt.Errorf("database: unable to parse array; unexpected %q at offset %d", src[i], i)
					}
					if bytes.Equal(elem, []byte("NULL")) {
						elem = nil
					}
					elems = append(elems, elem)
					break Element
				}
			}
		}
	}

	for i < len(src) {
		if bytes.HasPrefix(src[i:], del) && depth > 0 {
			dims[depth-1]++
			i += len(del)
			goto Element
		} else if src[i] == '}' && depth > 0 {
			dims[depth-1]++
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("database: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}

Close:
	for i < len(src) {
		if src[i] == '}' && depth > 0 {
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("database: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}
	if depth > 0 {
		err = fmt.Errorf("database: unable to parse array; expected %q at offset %d", '}', i)
	}
	if err == nil {
		for _, d := range dims {
			if (len(elems) % d) != 0 {
				err = fmt.Errorf("database: multidimensional arrays must have elements with matching dimensions")
			}
		}
	}
	return
}

func scanLinearArray(src, del []byte, typ string) (elems [][]byte, err error) {
	dims, elems, err := parseArray(src, del)
	if err != nil {
		return nil, err
	}
	if len(dims) > 1 {
		return nil, fmt.Errorf("database: cannot convert ARRAY%s to %s", strings.Replace(fmt.Sprint(dims), " ", "][", -1), typ)
	}
	return elems, err
}
