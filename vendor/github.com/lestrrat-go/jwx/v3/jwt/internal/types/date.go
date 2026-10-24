package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/jwx/v3/internal/json"
	"github.com/lestrrat-go/jwx/v3/internal/tokens"
)

const (
	DefaultPrecision uint32 = 0 // second level
	MaxPrecision     uint32 = 9 // nanosecond level
)

var Pedantic atomic.Uint32
var ParsePrecision atomic.Uint32
var FormatPrecision atomic.Uint32

// NumericDate represents the date format used in the 'nbf' claim
type NumericDate struct {
	time.Time
}

func (n *NumericDate) Get() time.Time {
	if n == nil {
		return (time.Time{}).UTC()
	}
	return n.Time
}

// validateNumericDateTime checks that Unix seconds and time.Time agree about
// whether a date is before January 1, 1970 UTC. Negative seconds must be
// before that instant; zero or positive seconds must be on or after it.
// time.Unix(math.MaxInt64, 0) breaks this rule: adding Go's internal time
// offset overflows, so Unix() stays positive while comparisons put the date
// before 1970.
func validateNumericDateTime(t time.Time) error {
	seconds := t.Unix()
	if (seconds < 0) != t.Before(time.Unix(0, 0)) {
		return fmt.Errorf(`NumericDate %d is out of range for time.Time`, seconds)
	}
	return nil
}

func intToTime(v any, t *time.Time) bool {
	var n int64
	switch x := v.(type) {
	case int64:
		n = x
	case int32:
		n = int64(x)
	case int16:
		n = int64(x)
	case int8:
		n = int64(x)
	case int:
		n = int64(x)
	default:
		return false
	}

	*t = time.Unix(n, 0)
	return true
}

const decimalDigits = "0123456789"

// parseEpochSeconds parses a decimal number of seconds since the epoch: an
// optional '-', one or more digits, and an optional fraction. The sign applies
// to the whole number, fraction included. Fractional digits beyond
// ParsePrecision are dropped, rounding toward negative infinity the same way
// NumericDate.String() and time.Unix do.
func parseEpochSeconds(x string) (time.Time, error) {
	digits := strings.TrimPrefix(x, "-")
	negative := len(digits) < len(x)
	whole, fractional, _ := strings.Cut(digits, string(tokens.Period))
	if whole == "" || strings.TrimLeft(whole, decimalDigits) != "" || strings.TrimLeft(fractional, decimalDigits) != "" {
		return time.Time{}, fmt.Errorf(`invalid number of seconds %q`, x)
	}

	// Parse the sign together with the whole part, so that the most
	// negative int64 still fits
	sec, err := strconv.ParseInt(x[:len(x)-len(digits)+len(whole)], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf(`failed to parse whole value %q: %w`, whole, err)
	}

	kept := fractional
	if precision := int(ParsePrecision.Load()); len(kept) > precision {
		kept = kept[:precision]
	}
	// unit ends up as the value of the last kept digit, in nanoseconds
	unit := int64(time.Second)
	var nsec int64
	for i := range len(kept) {
		unit /= 10
		nsec += int64(kept[i]-'0') * unit
	}

	if !negative {
		return time.Unix(sec, nsec).UTC(), nil
	}
	// Dropping digits moved a negative value toward zero, so move it one
	// unit back down. "-0" parses as 0, so the sign is applied through nsec,
	// and time.Unix carries it into the seconds.
	if strings.Trim(fractional[len(kept):], "0") != "" {
		nsec += unit
	}
	return time.Unix(sec, -nsec).UTC(), nil
}

func parseNumericString(x string) (time.Time, error) {
	t, err := parseEpochSeconds(x)
	if err == nil || Pedantic.Load() == 1 {
		return t, err
	}

	// This is an escape hatch for non-conformant providers
	// that gives us RFC3339 instead of epoch time
	tv, err := time.Parse(time.RFC3339, x)
	if err != nil {
		return time.Time{}, fmt.Errorf(`value is not number of seconds since the epoch, and attempt to parse it as RFC3339 timestamp failed: %w`, err)
	}
	return tv, nil
}

func (n *NumericDate) Accept(v any) error {
	var t time.Time
	switch x := v.(type) {
	case float32:
		tv, err := parseNumericString(fmt.Sprintf(`%.9f`, x))
		if err != nil {
			return fmt.Errorf(`failed to accept float32 %.9f: %w`, x, err)
		}
		t = tv
	case float64:
		tv, err := parseNumericString(fmt.Sprintf(`%.9f`, x))
		if err != nil {
			return fmt.Errorf(`failed to accept float64 %.9f: %w`, x, err)
		}
		t = tv
	case json.Number:
		tv, err := parseNumericString(x.String())
		if err != nil {
			return fmt.Errorf(`failed to accept json.Number %q: %w`, x.String(), err)
		}
		t = tv
	case string:
		tv, err := parseNumericString(x)
		if err != nil {
			return fmt.Errorf(`failed to accept string %q: %w`, x, err)
		}
		t = tv
	case time.Time:
		t = x
	default:
		if !intToTime(v, &t) {
			return fmt.Errorf(`invalid type %T`, v)
		}
	}
	if err := validateNumericDateTime(t); err != nil {
		return err
	}
	n.Time = t.UTC()
	return nil
}

func (n NumericDate) String() string {
	formatPrecision := FormatPrecision.Load()
	if formatPrecision == 0 {
		return strconv.FormatInt(n.Unix(), 10)
	}

	// Work from seconds and nanoseconds separately: UnixNano() only covers
	// the years 1678 through 2262, while a NumericDate can hold any
	// time.Time.
	//
	// Unix() floors, so the nanoseconds are always added to the seconds,
	// even before 1970. The fraction is floored to the requested digits
	// too, which keeps precision N consistent with precision 0.
	sec := n.Unix()
	scale := int64(1)
	for range formatPrecision {
		scale *= 10
	}
	frac := int64(n.Nanosecond()) / (int64(time.Second) / scale)
	if frac == 0 {
		// Whole seconds are written as an integer, the same as at
		// precision 0
		return strconv.FormatInt(sec, 10)
	}

	var buf [32]byte
	b := buf[:0]
	if sec < 0 {
		// sec + frac/scale is a negative number whose whole part is
		// -(sec+1) and whose fraction is scale-frac. For example,
		// sec=-2 and frac=5 at precision 1 is written as "-1.5".
		b = append(b, '-')
		b = strconv.AppendInt(b, -(sec + 1), 10)
		frac = scale - frac
	} else {
		b = strconv.AppendInt(b, sec, 10)
	}
	b = append(b, tokens.Period)

	// Trim trailing zeros: at precision 3, 0.100 is written as "0.1".
	// frac is not zero here, so the loop stops at the last nonzero digit.
	width := int(formatPrecision)
	for frac%10 == 0 {
		frac /= 10
		width--
	}

	// Write width digits, keeping leading zeros: at precision 3, 0.001
	// is written as "0.001"
	var digits [MaxPrecision]byte
	for i := width - 1; i >= 0; i-- {
		digits[i] = byte('0' + frac%10)
		frac /= 10
	}
	b = append(b, digits[:width]...)
	return string(b)
}

// MarshalJSON translates from internal representation to JSON NumericDate
// See https://tools.ietf.org/html/rfc7519#page-6
func (n *NumericDate) MarshalJSON() ([]byte, error) {
	if n.IsZero() {
		return json.Marshal(nil)
	}

	return json.Marshal(n.String())
}

// isPlainDecimal reports whether data is a JSON number without an
// exponent: an optional '-', a whole part without leading zeros, and an
// optional '.' followed by at least one digit
func isPlainDecimal(data []byte) bool {
	whole, fractional, hasPeriod := bytes.Cut(bytes.TrimPrefix(data, []byte{'-'}), []byte{tokens.Period})
	if len(whole) == 0 || (len(whole) > 1 && whole[0] == '0') || (hasPeriod && len(fractional) == 0) {
		return false
	}
	return len(bytes.TrimLeft(whole, decimalDigits)) == 0 && len(bytes.TrimLeft(fractional, decimalDigits)) == 0
}

func (n *NumericDate) UnmarshalJSON(data []byte) error {
	// Numbers without an exponent are parsed from their digits. Decoding
	// them into a float64 would keep only about 16 significant digits,
	// which at current timestamps is about 7 fractional digits.
	if isPlainDecimal(data) {
		t, err := parseEpochSeconds(string(data))
		if err != nil {
			return fmt.Errorf(`invalid value for NumericDate: %w`, err)
		}
		// Accept applies the checks that every other input goes through
		if err := n.Accept(t); err != nil {
			return fmt.Errorf(`invalid value for NumericDate: %w`, err)
		}
		return nil
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf(`failed to unmarshal date: %w`, err)
	}

	var n2 NumericDate
	if err := n2.Accept(v); err != nil {
		return fmt.Errorf(`invalid value for NumericDate: %w`, err)
	}
	*n = n2
	return nil
}
