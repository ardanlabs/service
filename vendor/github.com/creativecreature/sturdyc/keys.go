package sturdyc

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func handleSlice(v reflect.Value) string {
	// If the value is a pointer to a slice, get the actual slice
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Len() < 1 {
		return "empty"
	}

	var sliceStrings []string
	for i := 0; i < v.Len(); i++ {
		sliceStrings = append(sliceStrings, fmt.Sprintf("%v", v.Index(i).Interface()))
	}

	return strings.Join(sliceStrings, ",")
}

func extractPermutation(cacheKey string) string {
	idIndex := strings.LastIndex(cacheKey, "ID-")

	// "ID-" not found, return the original cache key.
	if idIndex == -1 {
		return cacheKey
	}

	// Find the last "-" before "ID-" to ensure we include "ID-" in the result
	lastDashIndex := strings.LastIndex(cacheKey[:idIndex], "-")
	// "-" not found before "ID-", return original string
	if lastDashIndex == -1 {
		return cacheKey
	}

	return cacheKey[:lastDashIndex+1]
}

func (c *Client[T]) relativeTime(t time.Time) string {
	now := c.clock.Now().Truncate(c.keyTruncation)
	target := t.Truncate(c.keyTruncation)
	var diff time.Duration
	var direction string
	if target.After(now) {
		diff = target.Sub(now)
		direction = "(+)" // Time is in the future
	} else {
		diff = now.Sub(target)
		direction = "(-)" // Time is in the past
	}
	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60
	return fmt.Sprintf("%s%dh%02dm%02ds", direction, hours, minutes, seconds)
}

// handleTime turns the time.Time into an epoch string.
func (c *Client[T]) handleTime(v reflect.Value) string {
	if timestamp, ok := v.Interface().(time.Time); ok {
		if !timestamp.IsZero() {
			if c.useRelativeTimeKeyFormat {
				return c.relativeTime(timestamp)
			}
			return strconv.FormatInt(timestamp.Unix(), 10)
		}
	}

	return "empty-time"
}

// PermutatedKey is a helper function for creating a cache key from a struct of
// options. Passing anything but a struct for "permutationStruct" will result
// in a panic.
func (c *Client[T]) PermutatedKey(prefix string, permutationStruct interface{}) string {
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteString("-")

	// Get the value of the interface
	v := reflect.ValueOf(permutationStruct)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("val must be a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// Check if the field is exported
		if !field.CanInterface() {
			continue // Skip unexported fields
		}

		if i > 0 {
			sb.WriteString("-")
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				sb.WriteString("nil")
				continue
			}
			// If it's not nil we'll dereference the pointer to handle its value.
			field = field.Elem()
		}

		//nolint:exhaustive // We only need special logic for slices and time.Time values.
		switch field.Kind() {
		case reflect.Slice:
			if field.IsNil() {
				sb.WriteString("nil")
			} else {
				sliceString := handleSlice(field)
				sb.WriteString(sliceString)
			}
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(time.Time{}) {
				sb.WriteString(c.handleTime(field))
				continue
			}
			sb.WriteString(fmt.Sprintf("%v", field.Interface()))
		default:
			sb.WriteString(fmt.Sprintf("%v", field.Interface()))
		}
	}

	return sb.String()
}

// BatchKeyFn provides a function for that can be used in conjunction with "GetFetchBatch".
// It takes in a prefix, and returns a function that will append an ID suffix for each item.
func (c *Client[T]) BatchKeyFn(prefix string) KeyFn {
	return func(id string) string {
		return fmt.Sprintf("%s-ID-%s", prefix, id)
	}
}

// PermutatedBatchKeyFn provides a function that can be used in conjunction
// with GetFetchBatch. It takes a prefix, and a struct where the fields are
// concatenated with the id in order to make a unique key. Passing anything but
// a struct for "permutationStruct" will result in a panic. This function is useful
// when the id isn't enough in itself to uniquely identify a record.
func (c *Client[T]) PermutatedBatchKeyFn(prefix string, permutationStruct interface{}) KeyFn {
	return func(id string) string {
		key := c.PermutatedKey(prefix, permutationStruct)
		return fmt.Sprintf("%s-ID-%s", key, id)
	}
}
