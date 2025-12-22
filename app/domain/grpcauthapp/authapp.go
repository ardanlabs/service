// Package grpcauthapp maintains the gRPC service for authentication.
package grpcauthapp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/app/sdk/mid"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// App represents the gRPC service for authentication.
type App struct {
	UnimplementedAuthServer
	log     *logger.Logger
	auth    *auth.Auth
	lis     net.Listener
	userBus userbus.ExtBusiness
	gs      *grpc.Server
}

func newApp(cfg Config) *App {
	return &App{
		log:     cfg.Log,
		auth:    cfg.Auth,
		lis:     cfg.Listener,
		userBus: cfg.UserBus,
	}
}

// Shutdown stops the gRPC service.
func (a *App) Shutdown() {
	a.log.Info(context.Background(), "shutdown", "status", "gRPC server stopped")
	a.gs.GracefulStop()
}

// Token generates a token for a given key ID.
func (a *App) Token(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	kid := req.GetKid()
	if kid == "" {
		return nil, status.Error(codes.InvalidArgument, "kid is required")
	}

	// For simplicity, we're using a dummy claims structure here.
	// In a real implementation, you would extract claims from the context or request.
	claims := auth.Claims{
		Roles: []string{"user"},
	}

	token, err := a.auth.GenerateToken(kid, claims)
	if err != nil {
		a.log.Error(ctx, "token", "err", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	trb := TokenResponse_builder{
		Token: proto.String(token),
	}

	return trb.Build(), nil
}

// Authenticate validates a bearer token.
func (a *App) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
	reqToken := req.GetToken()
	if reqToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	a.log.Info(ctx, "authenticate", "method", "grpc")

	claims, err := a.auth.Authenticate(ctx, reqToken)
	if err != nil {
		a.log.Error(ctx, "authenticate", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	arb := AuthenticateResponse_builder{
		UserId:  proto.String(claims.Subject),
		Subject: proto.String(claims.Subject),
		Roles:   claims.Roles,
	}

	return arb.Build(), nil
}

// Authorize checks if a user is authorized for a specific rule.
func (a *App) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	reqUserID := req.GetUserId()
	if reqUserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	reqRule := req.GetRule()
	if reqRule == "" {
		return nil, status.Error(codes.InvalidArgument, "rule is required")
	}

	reqClaims := req.GetClaims()

	// Convert gRPC request to auth client format
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        reqClaims.GetId(),
			Issuer:    reqClaims.GetIssuer(),
			Subject:   reqClaims.GetSubject(),
			Audience:  reqClaims.GetAudience(),
			ExpiresAt: int64ToND(reqClaims.GetExpiresAt()),
			NotBefore: int64ToND(reqClaims.GetNotBefore()),
			IssuedAt:  int64ToND(reqClaims.GetIssuedAt()),
		},
		Roles: reqClaims.GetRoles(),
	}

	userID, err := uuid.Parse(reqUserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	err = a.auth.Authorize(ctx, claims, userID, reqRule)
	if err != nil {
		if errors.Is(err, auth.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, "not authorized")
		}

		a.log.Error(ctx, "authorize", "err", err)
		return nil, status.Error(codes.Internal, "authorization failed")
	}

	return &AuthorizeResponse{}, nil
}

// =============================================================================

func (a *App) authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case "/auth.Auth/Token":
		return a.authorize(ctx, req, info, handler)

	case "/auth.Auth/Authenticate":
		return a.authenticate(ctx, req, info, handler)

	default:
		return handler(ctx, req)
	}
}

func (a *App) authorize(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("unauthorized: no authorization header")
	}

	ctx, err := mid.HandleAuthorization(ctx, authHeaders[0], a.userBus, a.auth)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return handler(ctx, req)
}

func (a *App) authenticate(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("unauthorized: no authorization header")
	}

	ctx, err := mid.HandleAuthentication(ctx, a.auth, authHeaders[0])
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return handler(ctx, req)
}

func int64ToND(in int64) *jwt.NumericDate {
	round, frac := math.Modf(float64(in))
	return jwt.NewNumericDate(time.Unix(int64(round), int64(frac*1e9)))
}
