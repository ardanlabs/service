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

package dbarray

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const (
	infinityTSEnabledAlready        = "database: infinity timestamp enabled already"
	infinityTSNegativeMustBeSmaller = "database: infinity timestamp: negative value must be smaller (before) than positive"
)

var infinityTSEnabled = false
var infinityTSNegative time.Time
var infinityTSPositive time.Time

type parameterStatus struct {
	// server version in the same format as server_version_num, or 0 if unavailable.
	serverVersion int
}

// EnableInfinityTS controls the handling of Postgres' "-infinity" and
// "infinity" "timestamp"s.
//
// If EnableInfinityTS is not called, "-infinity" and "infinity" will return
// []byte("-infinity") and []byte("infinity") respectively, and potentially
// cause error "sql: Scan error on column index 0: unsupported driver -> Scan
// pair: []uint8 -> *time.Time", when scanning into a time.Time value.
//
// Once EnableInfinityTS has been called, all connections created using this
// driver will decode Postgres' "-infinity" and "infinity" for "timestamp",
// "timestamp with time zone" and "date" types to the predefined minimum and
// maximum times, respectively.  When encoding time.Time values, any time which
// equals or precedes the predefined minimum time will be encoded to
// "-infinity".  Any values at or past the maximum time will similarly be
// encoded to "infinity".
//
// If EnableInfinityTS is called with negative >= positive, it will panic.
// Calling EnableInfinityTS after a connection has been established results in
// undefined behavior.  If EnableInfinityTS is called more than once, it will
// panic.
func EnableInfinityTS(negative time.Time, positive time.Time) {
	if infinityTSEnabled {
		panic(infinityTSEnabledAlready)
	}
	if !negative.Before(positive) {
		panic(infinityTSNegativeMustBeSmaller)
	}
	infinityTSEnabled = true
	infinityTSNegative = negative
	infinityTSPositive = positive
}

func encode(parameterStatus *parameterStatus, x interface{}, oid int) []byte {
	const oidBytea = 17

	switch v := x.(type) {
	case int64:
		return strconv.AppendInt(nil, v, 10)
	case float64:
		return strconv.AppendFloat(nil, v, 'f', -1, 64)
	case []byte:
		if oid == oidBytea {
			return encodeBytea(parameterStatus.serverVersion, v)
		}

		return v
	case string:
		if oid == oidBytea {
			return encodeBytea(parameterStatus.serverVersion, []byte(v))
		}

		return []byte(v)
	case bool:
		return strconv.AppendBool(nil, v)
	case time.Time:
		return formatTS(v)

	default:
		errorf("encode: unknown type for %T", v)
	}

	panic("not reached")
}

// formatTS formats t into a format postgres understands.
func formatTS(t time.Time) []byte {
	if infinityTSEnabled {
		// t <= -infinity : ! (t > -infinity)
		if !t.After(infinityTSNegative) {
			return []byte("-infinity")
		}
		// t >= infinity : ! (!t < infinity)
		if !t.Before(infinityTSPositive) {
			return []byte("infinity")
		}
	}
	return formatTimestamp(t)
}

// formatTimestamp formats t into Postgres' text format for timestamps.
func formatTimestamp(t time.Time) []byte {
	// Need to send dates before 0001 A.D. with " BC" suffix, instead of the
	// minus sign preferred by Go.
	// Beware, "0000" in ISO is "1 BC", "-0001" is "2 BC" and so on
	bc := false
	if t.Year() <= 0 {
		// flip year sign, and add 1, e.g: "0" will be "1", and "-10" will be "11"
		t = t.AddDate((-t.Year())*2+1, 0, 0)
		bc = true
	}
	b := []byte(t.Format("2006-01-02 15:04:05.999999999Z07:00"))

	_, offset := t.Zone()
	offset %= 60
	if offset != 0 {
		// RFC3339Nano already printed the minus sign
		if offset < 0 {
			offset = -offset
		}

		b = append(b, ':')
		if offset < 10 {
			b = append(b, '0')
		}
		b = strconv.AppendInt(b, int64(offset), 10)
	}

	if bc {
		b = append(b, " BC"...)
	}
	return b
}

func errorf(s string, args ...interface{}) {
	panic(fmt.Errorf("pq: %s", fmt.Sprintf(s, args...)))
}

// Parse a bytea value received from the server.  Both "hex" and the legacy
// "escape" format are supported.
func parseBytea(s []byte) (result []byte, err error) {
	if len(s) >= 2 && bytes.Equal(s[:2], []byte("\\x")) {
		// bytea_output = hex
		s = s[2:] // trim off leading "\\x"
		result = make([]byte, hex.DecodedLen(len(s)))
		_, err := hex.Decode(result, s)
		if err != nil {
			return nil, err
		}
	} else {
		// bytea_output = escape
		for len(s) > 0 {
			if s[0] == '\\' {
				// escaped '\\'
				if len(s) >= 2 && s[1] == '\\' {
					result = append(result, '\\')
					s = s[2:]
					continue
				}

				// '\\' followed by an octal number
				if len(s) < 4 {
					return nil, fmt.Errorf("invalid bytea sequence %v", s)
				}
				r, err := strconv.ParseUint(string(s[1:4]), 8, 8)
				if err != nil {
					return nil, fmt.Errorf("could not parse bytea value: %s", err.Error())
				}
				result = append(result, byte(r))
				s = s[4:]
			} else {
				// We hit an unescaped, raw byte.  Try to read in as many as
				// possible in one go.
				i := bytes.IndexByte(s, '\\')
				if i == -1 {
					result = append(result, s...)
					break
				}
				result = append(result, s[:i]...)
				s = s[i:]
			}
		}
	}

	return result, nil
}

func encodeBytea(serverVersion int, v []byte) (result []byte) {
	if serverVersion >= 90000 {
		// Use the hex format if we know that the server supports it
		result = make([]byte, 2+hex.EncodedLen(len(v)))
		result[0] = '\\'
		result[1] = 'x'
		hex.Encode(result[2:], v)
	} else {
		// .. or resort to "escape"
		for _, b := range v {
			if b == '\\' {
				result = append(result, '\\', '\\')
			} else if b < 0x20 || b > 0x7e {
				result = append(result, []byte(fmt.Sprintf("\\%03o", b))...)
			} else {
				result = append(result, b)
			}
		}
	}

	return result
}
