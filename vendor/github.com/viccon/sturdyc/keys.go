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

// PermutatedKey takes a prefix and a struct where the fields are concatenated
// in order to create a unique cache key. Passing anything but a struct for
// "permutationStruct" will result in a panic. The cache will only use the
// EXPORTED fields of the struct to construct the key. The permutation struct
// should be FLAT, with no nested structs. The fields can be any of the basic
// types, as well as slices and time.Time values.
//
// Parameters:
//
//	prefix - The prefix for the cache key.
//	permutationStruct - A struct whose fields are concatenated to form a unique cache key.
//	Only exported fields are used.
//
// Returns:
//
//	A string to be used as the cache key.
//
// Example usage:
//
//	type queryParams struct {
//		City               string
//		Country            string
//	}
//	params := queryParams{"Stockholm", "Sweden"}
//	key := c.PermutatedKey("prefix",, params) // prefix-Stockholm-Sweden-1
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

		// Check if the field is exported, and if so skip it.
		if !field.CanInterface() {
			message := fmt.Sprintf(
				"sturdyc: permutationStruct contains unexported field: %s which won't be part of the cache key",
				v.Type().Field(i).Name,
			)
			c.log.Warn(message)
			continue
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

// BatchKeyFn provides a function that can be used in conjunction with
// "GetOrFetchBatch". It takes in a prefix and returns a function that will use
// the prefix, add a -ID- separator, and then append the ID as a suffix for
// each item.
//
// Parameters:
//
//	prefix - The prefix to be used for each cache key.
//
// Returns:
//
//	A function that takes an ID and returns a cache key string with the given prefix and ID.
//
// Example usage:
//
//	fn := c.BatchKeyFn("some-prefix")
//	key := fn("1234") // some-prefix-ID-1234
func (c *Client[T]) BatchKeyFn(prefix string) KeyFn {
	return func(id string) string {
		return fmt.Sprintf("%s-ID-%s", prefix, id)
	}
}

// PermutatedBatchKeyFn provides a function that can be used in conjunction
// with GetOrFetchBatch. It takes a prefix and a struct where the fields are
// concatenated with the ID in order to make a unique cache key. Passing
// anything but a struct for "permutationStruct" will result in a panic. The
// cache will only use the EXPORTED fields of the struct to construct the key.
// The permutation struct should be FLAT, with no nested structs. The fields
// can be any of the basic types, as well as slices and time.Time values.
//
// Parameters:
//
//	prefix - The prefix for the cache key.
//	permutationStruct - A struct whose fields are concatenated to form a unique cache key. Only exported fields are used.
//
// Returns:
//
//	A function that takes an ID and returns a cache key string with the given prefix, permutation struct fields, and ID.
//
// Example usage:
//
//	type queryParams struct {
//		City               string
//		Country            string
//	}
//	params := queryParams{"Stockholm", "Sweden"}
//	cacheKeyFunc := c.PermutatedBatchKeyFn("prefix", params)
//	key := cacheKeyFunc("1") // prefix-Stockholm-Sweden-ID-1
func (c *Client[T]) PermutatedBatchKeyFn(prefix string, permutationStruct interface{}) KeyFn {
	return func(id string) string {
		key := c.PermutatedKey(prefix, permutationStruct)
		return fmt.Sprintf("%s-ID-%s", key, id)
	}
}
