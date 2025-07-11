// Package unittest provides support for excuting unit test logic.
package unittest

import (
	"context"
	"testing"
)

type testOption struct {
	skip    bool
	skipMsg string
}

type OptionFunc func(*testOption)

// WithSkip can be used to skip running a test.
func WithSkip(skip bool, msg string) OptionFunc {
	return func(to *testOption) {
		to.skip = skip
		to.skipMsg = msg
	}
}

// Run performs the actual test logic based on the table data.
func Run(t *testing.T, table []Table, testName string, options ...OptionFunc) {
	to := new(testOption)
	for _, f := range options {
		f(to)
	}

	if to.skip {
		t.Skipf("%v: %v", testName, to.skipMsg)
	}

	for _, tt := range table {
		f := func(t *testing.T) {
			gotResp := tt.ExcFunc(context.Background())

			diff := tt.CmpFunc(gotResp, tt.ExpResp)
			if diff != "" {
				t.Log("DIFF")
				t.Logf("%s", diff)
				t.Log("GOT")
				t.Logf("%#v", gotResp)
				t.Log("EXP")
				t.Logf("%#v", tt.ExpResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
