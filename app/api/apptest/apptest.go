// Package apptest contains supporting code for running app layer tests.
package apptest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-json-experiment/json"
)

// AppTable represent fields needed for running an app test.
type AppTable struct {
	Name       string
	URL        string
	Token      string
	Method     string
	StatusCode int
	Model      any
	Resp       any
	ExpResp    any
	CmpFunc    func(got any, exp any) string
}

// =============================================================================

// AppTest contains functions for executing an app test.
type AppTest struct {
	handler http.Handler
}

func New(handler http.Handler) *AppTest {
	return &AppTest{
		handler: handler,
	}
}

// Test performs the actual test logic based on the table data.
func (at *AppTest) Test(t *testing.T, table []AppTable, testName string) {
	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(tt.Method, tt.URL, nil)
			w := httptest.NewRecorder()

			if tt.Model != nil {
				var b bytes.Buffer
				if err := json.MarshalWrite(&b, tt.Model, json.FormatNilSliceAsNull(true)); err != nil {
					t.Fatalf("Should be able to marshal the model : %s", err)
				}

				r = httptest.NewRequest(tt.Method, tt.URL, &b)
			}

			r.Header.Set("Authorization", "Bearer "+tt.Token)
			at.handler.ServeHTTP(w, r)

			if w.Code != tt.StatusCode {
				t.Fatalf("%s: Should receive a status code of %d for the response : %d", tt.Name, tt.StatusCode, w.Code)
			}

			if tt.StatusCode == http.StatusNoContent {
				return
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.Resp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := tt.CmpFunc(tt.Resp, tt.ExpResp)
			if diff != "" {
				t.Log("DIFF")
				t.Logf("%s", diff)
				t.Log("GOT")
				t.Logf("%#v", tt.Resp)
				t.Log("EXP")
				t.Logf("%#v", tt.ExpResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
