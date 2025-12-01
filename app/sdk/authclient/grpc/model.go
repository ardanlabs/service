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
	claims := grpcauthapp.Claims{
		Id:        auth.Claims.ID,
		Issuer:    auth.Claims.Issuer,
		Subject:   auth.Claims.Subject,
		Audience:  auth.Claims.Audience,
		ExpiresAt: auth.Claims.ExpiresAt.Unix(),
		NotBefore: auth.Claims.NotBefore.Unix(),
		IssuedAt:  auth.Claims.IssuedAt.Unix(),
		Roles:     auth.Claims.Roles,
	}

	req := grpcauthapp.AuthorizeRequest{
		UserId: auth.UserID.String(),
		Claims: &claims,
		Rule:   auth.Rule,
	}

	return &req
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
