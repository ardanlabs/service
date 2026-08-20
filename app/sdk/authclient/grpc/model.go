package grpc

import (
	"math"
	"time"

	"github.com/ardanlabs/service/app/domain/grpcauthapp"
	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/authclient"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func authorizeRequestToGRPC(auth authclient.Authorize) *grpcauthapp.AuthorizeRequest {
	cb := grpcauthapp.Claims_builder{
		Id:        new(auth.Claims.ID),
		Issuer:    new(auth.Claims.Issuer),
		Subject:   new(auth.Claims.Subject),
		Audience:  auth.Claims.Audience,
		ExpiresAt: new(auth.Claims.ExpiresAt.Unix()),
		NotBefore: new(auth.Claims.NotBefore.Unix()),
		IssuedAt:  new(auth.Claims.IssuedAt.Unix()),
		Roles:     auth.Claims.Roles,
	}
	claims := cb.Build()

	arb := grpcauthapp.AuthorizeRequest_builder{
		UserId: new(auth.UserID.String()),
		Rule:   new(auth.Rule),
		Claims: claims,
	}

	req := arb.Build()
	return req
}

func authenticateRespFromGRPC(a *grpcauthapp.AuthenticateResponse) (authclient.AuthenticateResp, error) {
	userID := a.GetUserId()
	uID, err := uuid.Parse(userID)
	if err != nil {
		return authclient.AuthenticateResp{}, err
	}

	c := a.GetClaims()
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        c.GetId(),
			Issuer:    c.GetIssuer(),
			Subject:   c.GetSubject(),
			Audience:  c.GetAudience(),
			ExpiresAt: int64ToND(c.GetExpiresAt()),
			NotBefore: int64ToND(c.GetNotBefore()),
			IssuedAt:  int64ToND(c.GetIssuedAt()),
		},
		Roles: c.GetRoles(),
	}

	resp := authclient.AuthenticateResp{
		UserID: uID,
		Claims: claims,
	}

	return resp, nil
}

func int64ToND(in int64) *jwt.NumericDate {
	round, frac := math.Modf(float64(in))
	return jwt.NewNumericDate(time.Unix(int64(round), int64(frac*1e9)))
}
