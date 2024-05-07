package sturdyc

import "fmt"

// safeGo is a helper that prevents panics in any of the goroutines
// that are running in the background from crashing the process.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				//nolint:forbidigo // This should never panic, but we want to log it if it does.
				fmt.Println(err)
			}
		}()
		fn()
	}()
}
