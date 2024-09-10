// Package apitest provides support for excuting api test logic.
package apitest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"testing"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/dbtest"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/golang-jwt/jwt/v4"
)

// Test contains functions for executing an api test.
type Test struct {
	DB   *dbtest.Database
	Auth *auth.Auth
	mux  http.Handler
}

// Run performs the actual test logic based on the table data.
func (at *Test) Run(t *testing.T, table []Table, testName string) {
	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(tt.Method, tt.URL, nil)
			w := httptest.NewRecorder()

			if tt.Input != nil {
				d, err := json.Marshal(tt.Input)
				if err != nil {
					t.Fatalf("Should be able to marshal the model : %s", err)
				}

				r = httptest.NewRequest(tt.Method, tt.URL, bytes.NewBuffer(d))
			}

			r.Header.Set("Authorization", "Bearer "+tt.Token)
			at.mux.ServeHTTP(w, r)

			if w.Code != tt.StatusCode {
				t.Fatalf("%s: Should receive a status code of %d for the response : %d", tt.Name, tt.StatusCode, w.Code)
			}

			if tt.StatusCode == http.StatusNoContent {
				return
			}

			if err := json.Unmarshal(w.Body.Bytes(), tt.GotResp); err != nil {
				t.Fatalf("Should be able to unmarshal the response : %s", err)
			}

			diff := tt.CmpFunc(tt.GotResp, tt.ExpResp)
			if diff != "" {
				t.Log("DIFF")
				t.Logf("%s", diff)
				t.Log("GOT")
				t.Logf("%#v", tt.GotResp)
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
func Token(userBus *userbus.Business, ath *auth.Auth, email string) string {
	addr, _ := mail.ParseAddress(email)

	dbUsr, err := userBus.QueryByEmail(context.Background(), *addr)
	if err != nil {
		return ""
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   dbUsr.ID.String(),
			Issuer:    ath.Issuer(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: role.ParseToString(dbUsr.Roles),
	}

	token, err := ath.GenerateToken(kid, claims)
	if err != nil {
		return ""
	}

	return token
}
