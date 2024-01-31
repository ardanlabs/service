package auth_test

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func Test_Auth(t *testing.T) {
	log, db, teardown := newUnit(t)
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			t.Error(string(debug.Stack()))
		}
		teardown()
	}()

	cfg := auth.Config{
		Log:       log,
		DB:        db,
		KeyLookup: &keyStore{},
		Issuer:    "service project",
	}
	a, err := auth.New(cfg)
	if err != nil {
		t.Fatalf("Should be able to create an authenticator: %s", err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleAdmin},
	}
	userID := uuid.MustParse(claims.Subject)

	token, err := a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %s", err)
	}

	parsedClaims, err := a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAdminOnly)
	if err != nil {
		t.Errorf("Should be able to authorize the RoleAdmin claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleUserOnly)
	if err == nil {
		t.Error("Should NOT be able to authorize the RoleUser claim")
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAdminOrSubject)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAdminOrSubject claim with RoleAdmin only : %s", err)
	}

	// -------------------------------------------------------------------------

	claims = auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleUser},
	}
	userID = uuid.MustParse(claims.Subject)

	token, err = a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %v", err)
	}

	parsedClaims, err = a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleUserOnly)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleUserOnly claim with RoleUser only : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAdminOnly)
	if err == nil {
		t.Error("Should NOT be able to authorize the RuleAdminOnly claim with RoleUser only")
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAdminOrSubject)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAdminOrSubject claim with RoleUser only : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAny)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAny any claim with RoleUser only : %s", err)
	}

	// -------------------------------------------------------------------------

	claims = auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleUser},
	}
	userID = uuid.MustParse("9e979baa-61c9-4b50-81f2-f216d53f5c15")

	token, err = a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %s", err)
	}

	parsedClaims, err = a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAdminOrSubject)
	if err == nil {
		t.Error("Should NOT be able to authorize the RuleAdminOrSubject claim with RoleUser only and different userID")
	}

	// -------------------------------------------------------------------------

	claims = auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleUser, user.RoleAdmin},
	}
	userID = uuid.MustParse("9e979baa-61c9-4b50-81f2-f216d53f5c15")

	token, err = a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %s", err)
	}

	parsedClaims, err = a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAny)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAny any claim with RoleUser and RoleAdmin : %s", err)
	}

	// -------------------------------------------------------------------------

	claims = auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleUser},
	}
	userID = uuid.MustParse("9e979baa-61c9-4b50-81f2-f216d53f5c15")

	token, err = a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %s", err)
	}

	parsedClaims, err = a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAny)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAny any claim with RoleUser only : %s", err)
	}

	// -------------------------------------------------------------------------

	claims = auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Subject:   "5cf37266-3473-4006-984f-9325122678b7",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []user.Role{user.RoleAdmin},
	}
	userID = uuid.MustParse("9e979baa-61c9-4b50-81f2-f216d53f5c15")

	token, err = a.GenerateToken(kid, claims)
	if err != nil {
		t.Fatalf("Should be able to generate a JWT : %s", err)
	}

	parsedClaims, err = a.Authenticate(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("Should be able to authenticate the claims : %s", err)
	}

	err = a.Authorize(context.Background(), parsedClaims, userID, auth.RuleAny)
	if err != nil {
		t.Errorf("Should be able to authorize the RuleAny any claim with RoleAdmin only : %s", err)
	}
}

func newUnit(t *testing.T) (*logger.Logger, *sqlx.DB, func()) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	// teardown is the function that should be invoked when the caller is done
	// with the database.
	teardown := func() {
		t.Helper()

		fmt.Println("******************** LOGS ********************")
		fmt.Print(buf.String())
		fmt.Println("******************** LOGS ********************")
	}

	return log, nil, teardown
}

type keyStore struct{}

func (ks *keyStore) PrivateKey(kid string) (string, error) {
	return privateKeyPEM, nil
}

func (ks *keyStore) PublicKey(kid string) (string, error) {
	return publicKeyPEM, nil
}

const (
	kid = "s4sKIjD9kIRjxs2tulPqGLdxSfgPErRN1Mu3Hd9k9NQ"

	privateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAvMAHb0IoLvoYuW2kA+LTmnk+hfnBq1eYIh4CT/rMPCxgtzjq
U0guQOMnLg69ydyA5uu37v6rbS1+stuBTEiMQl/bxAhgLkGrUhgpZ10Bt6GzSEgw
QNloZoGaxe4p20wMPpT4kcMKNHkQds3uONNcLxPUmfjbbH64g+seg28pbgQPwKFK
tF7bIsOBgz0g5Ptn5mrkdzqMPUSy9k9VCu+R42LH9c75JsRzz4FeN+VzwMAL6yQn
ZvOi7/zOgNyxeVia8XVKykrnhgcpiOn5oaLRBzQGN00Z7TuBRIfDJWU21qQN4Cq7
keZmMP4gqCVWjYneK4bzrG/+H2w9BJ2TsmMGvwIDAQABAoIBAFQmQKpHkmavNYql
6POaksBRwaA1YzSijr7XJizGIXvKRSwqgb2zdnuTSgpspAx09Dr/aDdy7rZ0DAJt
fk2mInINDottOIQm3txwzTS58GQQAT/+fxTKWJMqwPfxYFPWqbbU76T8kXYna0Gs
OcK36GdMrgIfQqQyMs0Na8MpMg1LmkAxuqnFCXS/NMyKl9jInaaTS+Kz+BSzUMGQ
zebfLFsf2N7sLZuimt9zlRG30JJTfBlB04xsYMo734usA2ITe8U0XqG6Og0qc6ev
6lsoM8hpvEUsQLcjQQ5up7xx3S2stZJ8o0X8GEX5qUMaomil8mZ7X5xOlEqf7p+v
lXQ46cECgYEA2lbZQON6l3ZV9PCn9j1rEGaXio3SrAdTyWK3D1HF+/lEjClhMkfC
XrECOZYj+fiI9n+YpSog+tTDF7FTLf7VP21d2gnhQN6KAXUnLIypzXxodcC6h+8M
ZGJh/EydLvC7nPNoaXx96bohxzS8hrOlOlkCbr+8gPYKf8qkbe7HyxECgYEA3U6e
x9g4FfTvI5MGrhp2BIzoRSn7HlNQzjJ71iMHmM2kBm7TsER8Co1PmPDrP8K/UyGU
Q25usTsPSrHtKQEV6EsWKaP/6p2Q82sDkT9bZlV+OjRvOfpdO5rP6Q95vUmMGWJ/
S6oimbXXL8p3gDafw3vC1PCAhoaxMnGyKuZwlM8CgYEAixT1sXr2dZMg8DV4mMfI
8pqXf+AVyhWkzsz+FVkeyAKiIrKdQp0peI5C/5HfevVRscvX3aY3efCcEfSYKt2A
07WEKkdO4LahrIoHGT7FT6snE5NgfwTMnQl6p2/aVLNun20CHuf5gTBbIf069odr
Af7/KLMkjfWs/HiGQ6zuQjECgYEAv+DIvlDz3+Wr6dYyNoXuyWc6g60wc0ydhQo0
YKeikJPLoWA53lyih6uZ1escrP23UOaOXCDFjJi+W28FR0YProZbwuLUoqDW6pZg
U3DxWDrL5L9NqKEwcNt7ZIDsdnfsJp5F7F6o/UiyOFd9YQb7YkxN0r5rUTg7Lpdx
eMyv0/UCgYEAhX9MPzmTO4+N8naGFof1o8YP97pZj0HkEvM0hTaeAQFKJiwX5ijQ
xumKGh//G0AYsjqP02ItzOm2mWnbI3FrNlKmGFvR6VxIZMOyXvpLofHucjJ5SWli
eYjPklKcXaMftt1FVO4n+EKj1k1+Tv14nytq/J5WN+r4FBlNEYj/6vg=
-----END RSA PRIVATE KEY-----
`
	publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvMAHb0IoLvoYuW2kA+LT
mnk+hfnBq1eYIh4CT/rMPCxgtzjqU0guQOMnLg69ydyA5uu37v6rbS1+stuBTEiM
Ql/bxAhgLkGrUhgpZ10Bt6GzSEgwQNloZoGaxe4p20wMPpT4kcMKNHkQds3uONNc
LxPUmfjbbH64g+seg28pbgQPwKFKtF7bIsOBgz0g5Ptn5mrkdzqMPUSy9k9VCu+R
42LH9c75JsRzz4FeN+VzwMAL6yQnZvOi7/zOgNyxeVia8XVKykrnhgcpiOn5oaLR
BzQGN00Z7TuBRIfDJWU21qQN4Cq7keZmMP4gqCVWjYneK4bzrG/+H2w9BJ2TsmMG
vwIDAQAB
-----END PUBLIC KEY-----
`
)
