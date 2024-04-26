package dbtest

import (
	"context"
	"testing"
)

// UnitTable represent fields needed for running an unit test.
type UnitTable struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}

// =============================================================================

// UnitTest performs the actual test logic based on the table data.
func UnitTest(t *testing.T, table []UnitTable, testName string) {
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
			got := tt.ExcFunc(ctx)

			diff := tt.CmpFunc(got, tt.ExpResp)
			if diff != "" {
				log(diff, got, tt.ExpResp)
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
