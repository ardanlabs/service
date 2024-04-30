// Package apitest provides support for excuting api test logic.
package apitest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/ardanlabs/service/app/api/auth"
	"github.com/ardanlabs/service/business/api/dbtest"
	"github.com/ardanlabs/service/business/domain/userbus/stores/userdb"
	"github.com/go-json-experiment/json"
	"github.com/golang-jwt/jwt/v4"
)

// Test contains functions for executing an api test.
type Test struct {
	DB      *dbtest.Database
	Auth    *auth.Auth
	handler http.Handler
}

// New constructs a Test value for running api tests.
func New(db *dbtest.Database, ath *auth.Auth, handler http.Handler) *Test {
	return &Test{
		DB:      db,
		Auth:    ath,
		handler: handler,
	}
}

// Run performs the actual test logic based on the table data.
func (at *Test) Run(t *testing.T, table []Table, testName string) {
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

// =============================================================================

// Token generates an authenticated token for a user.
func Token(db *dbtest.Database, ath *auth.Auth, email string) string {
	addr, _ := mail.ParseAddress(email)

	store := userdb.NewStore(db.Log, db.DB)
	dbUsr, err := store.QueryByEmail(context.Background(), *addr)
	if err != nil {
		return ""
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUsr.ID.String(),
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: dbUsr.Roles,
	}

	token, err := ath.GenerateToken(kid, claims)
	if err != nil {
		return ""
	}

	return token
}
