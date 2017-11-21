package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ardanlabs/service/internal/user"
)

// TestUsers is the entry point for the users
func TestUsers(t *testing.T) {
	t.Run("getUsers200Empty", getUsers200Empty)
	t.Run("postUser400", postUser400)
	// t.Run("getUser404", getUser404)
	// t.Run("getUser400", getUser400)
	// t.Run("deleteUser404", deleteUser404)
	// t.Run("putUser404", putUser404)
	// t.Run("crudUsers", crudUser)
}

// getUsers200Empty validates an empty users list can be retrieved with the endpoint.
func getUsers200Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/v1/users", nil)
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to fetch an empty list of users with the users endpoint.")
	{
		t.Log("\tTest 0:\tWhen fetching an empty user list.")
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tShould receive a status code of 200 for the response : %v", Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 200 for the response.", Success)

			recv := w.Body.String()
			resp := `[]`
			if resp != recv {
				t.Log("Got :", recv)
				t.Log("Want:", resp)
				t.Fatalf("\t%s\tShould get the expected result.", Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", Success)
		}
	}
}

// postUser400 validates a user can't be created with the endpoint
// unless a valid user document is submitted.
func postUser400(t *testing.T) {
	u := user.User{
		UserType: 1,
		LastName: "Kennedy",
		Email:    "bill@ardanstudios.com",
		Company:  "Ardan Labs",
	}

	body, _ := json.Marshal(&u)
	r := httptest.NewRequest("POST", "/v1/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	a.ServeHTTP(w, r)

	t.Log("Given the need to validate a new user can't be created with an invalid document.")
	{
		t.Log("\tTest 0:\tWhen using an incomplete user value.")
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tShould receive a status code of 400 for the response : %v", Failed, w.Code)
			}
			t.Logf("\t%s\tShould receive a status code of 400 for the response.", Success)

			recv := w.Body.String()
			resps := []string{
				`{
  "error": "field validation failure",
  "fields": [
    {
      "field_name": "FirstName",
      "error": "required"
    }
  ]
}`,
			}

			var found bool
			for _, resp := range resps {
				if resp == recv {
					found = true
					break
				}
			}

			if !found {
				t.Log("Got :", recv)
				t.Log("Want:", resps[0])
				t.Fatalf("\t%s\tShould get the expected result.", Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", Success)
		}
	}
}
