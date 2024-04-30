// Package unittest provides support for excuting unit test logic.
package unittest

import (
	"context"
	"testing"
)

// Table represent fields needed for running an unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}

// Run performs the actual test logic based on the table data.
func Run(t *testing.T, table []Table, testName string) {
	log := func(diff string, got any, exp any) {
		t.Log("DIFF")
		t.Logf("%s", diff)
		t.Log("GOT")
		t.Logf("%#v", got)
		t.Log("EXP")
		t.Logf("%#v", exp)
		t.Fatalf("Should get the expected response")
	}

	for _, tt := range table {
		f := func(t *testing.T) {
			ctx := context.Background()

			t.Log("Calling excFunc")
			gotResp := tt.ExcFunc(ctx)

			diff := tt.CmpFunc(gotResp, tt.ExpResp)
			if diff != "" {
				log(diff, gotResp, tt.ExpResp)
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
