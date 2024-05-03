// Package unitest provides support for excuting unit test logic.
package unitest

import (
	"context"
	"testing"
)

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
			gotResp := tt.ExcFunc(context.Background())

			diff := tt.CmpFunc(gotResp, tt.ExpResp)
			if diff != "" {
				log(diff, gotResp, tt.ExpResp)
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
