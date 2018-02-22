package cfg

import (
	"net/url"
	"time"
)

// c is the default Config used by Init and the package level funcs like
// String, MustString, and SetString.
var c Config

// Init populates the package's default Config and should be called only once.
// A Provider must be supplied which will return a map of key/value pairs to be
// loaded.
func Init(p Provider) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the provided configuration.
	m, err := p.Provide()
	if err != nil {
		return err
	}

	// Set it to the global instance.
	c.m = m

	return nil
}

// Log returns a string to help with logging the package's default Config. It
// excludes any values whose key contains the string "PASS".
func Log() string {
	return c.Log()
}

// String calls the default Config and returns the value of the given key as a
// string. It will return an error if key was not found.
func String(key string) (string, error) {
	return c.String(key)
}

// MustString calls the default Config and returns the value of the given key
// as a string, else it will panic if the key was not found.
func MustString(key string) string {
	return c.MustString(key)
}

// SetString adds or modifies the default Config for the specified key and
// value.
func SetString(key string, value string) {
	c.SetString(key, value)
}

// Int calls the Default config and returns the value of the given key as an
// int. It will return an error if the key was not found or the value
// can't be converted to an int.
func Int(key string) (int, error) {
	return c.Int(key)
}

// MustInt calls the default Config and returns the value of the given key as
// an int. It will panic if the key was not found or the value can't be
// converted to an int.
func MustInt(key string) int {
	return c.MustInt(key)
}

// SetInt adds or modifies the default Config for the specified key and value.
func SetInt(key string, value int) {
	c.SetInt(key, value)
}

// Time calls the default Config and returns the value of the given key as a
// Time. It will return an error if the key was not found or the value can't be
// converted to a Time.
func Time(key string) (time.Time, error) {
	return c.Time(key)
}

// MustTime calls the default Config ang returns the value of the given key as
// a Time. It will panic if the key was not found or the value can't be
// converted to a Time.
func MustTime(key string) time.Time {
	return c.MustTime(key)
}

// SetTime adds or modifies the default Config for the specified key and value.
func SetTime(key string, value time.Time) {
	c.SetTime(key, value)
}

// Bool calls the default Config and returns the bool value of a given key as a
// bool. It will return an error if the key was not found or the value can't be
// converted to a bool.
func Bool(key string) (bool, error) {
	return c.Bool(key)
}

// MustBool calls the default Config and returns the bool value of a given key
// as a bool. It will panic if the key was not found or the value can't be
// converted to a bool.
func MustBool(key string) bool {
	return c.MustBool(key)
}

// SetBool adds or modifies the default Config for the specified key and value.
func SetBool(key string, value bool) {
	c.SetBool(key, value)
}

// URL calls the default Config and returns the value of the given key as a
// URL. It will return an error if the key was not found or the value can't be
// converted to a URL.
func URL(key string) (*url.URL, error) {
	return c.URL(key)
}

// MustURL calls the default Config and returns the value of the given key as a
// URL. It will panic if the key was not found or the value can't be converted
// to a URL.
func MustURL(key string) *url.URL {
	return c.MustURL(key)
}

// SetURL adds or modifies the default Config for the specified key and value.
func SetURL(key string, value *url.URL) {
	c.SetURL(key, value)
}

// Duration calls the default Config and returns the value of the given key as a
// duration. It will return an error if the key was not found or the value can't be
// converted to a Duration.
func Duration(key string) (time.Duration, error) {
	return c.Duration(key)
}

// MustDuration calls the default Config and returns the value of the given
// key as a MustDuration. It will panic if the key was not found or the value
// can't be converted to a MustDuration.
func MustDuration(key string) time.Duration {
	return c.MustDuration(key)
}

// SetDuration adds or modifies the default Config for the specified key and value.
func SetDuration(key string, value time.Duration) {
	c.SetDuration(key, value)
}
