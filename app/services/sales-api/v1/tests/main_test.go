package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	var err error
	c, err = dbtest.StartDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dbtest.StopDB(c)

	m.Run()
}

type appTest struct {
	http.Handler
	userToken  string
	adminToken string
}

func (at *appTest) test(t *testing.T, table []tableData, testName string) {
	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			if tt.model != nil {
				var b bytes.Buffer
				if err := json.NewEncoder(&b).Encode(tt.model); err != nil {
					t.Fatalf("Should be able to marshal the model : %s", err)
				}

				r = httptest.NewRequest(tt.method, tt.url, &b)
			}

			r.Header.Set("Authorization", "Bearer "+tt.token)
			at.ServeHTTP(w, r)

			if w.Code != tt.statusCode {
				t.Fatalf("%s: Should receive a status code of %d for the response : %d", tt.name, tt.statusCode, w.Code)
			}

			if tt.statusCode == http.StatusNoContent {
				return
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.resp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := tt.cmpFunc(tt.resp, tt.expResp)
			if diff != "" {
				t.Log("GOT")
				t.Logf("%#v", tt.resp)
				t.Log("EXP")
				t.Logf("%#v", tt.expResp)
				t.Fatalf("Should get the expected response")
			}
		}

		t.Run(testName+"-"+tt.name, f)
	}
}
