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
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Config holds the dependencies for the gRPC service.
type Config struct {
	Log     *logger.Logger
	Auth    *auth.Auth
	UserBus userbus.ExtBusiness
	Host    string
	APIHost string
}

// Service represents the gRPC service for authentication.
type Service struct {
	UnimplementedAuthServer
	log     *logger.Logger
	auth    *auth.Auth
	userBus userbus.ExtBusiness
	host    string
	api     string
	gs      *grpc.Server
	lis     net.Listener
}

// New creates a new gRPC service.
func New(cfg Config) (*Service, error) {
	lis, err := net.Listen("tcp", cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	service := &Service{
		log:     cfg.Log,
		auth:    cfg.Auth,
		userBus: cfg.UserBus,
		host:    cfg.Host,
		api:     cfg.APIHost,
		lis:     lis,
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(service.authInterceptor),
	)
	service.gs = s

	RegisterAuthServer(s, service)

	reflection.Register(s)

	return service, nil
}

// Start starts the gRPC service.
func (s *Service) Start() error {
	s.log.Info(context.Background(), "startup", "status", "gRPC server started", "host", s.host)
	return s.gs.Serve(s.lis)
}

// Stop stops the gRPC service.
func (s *Service) Stop() {
	s.log.Info(context.Background(), "shutdown", "status", "gRPC server stopped", "host", s.host)
	s.gs.GracefulStop()
}

func (s *Service) authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	switch info.FullMethod {
	case "/auth.Auth/Token":
		return s.authorize(ctx, req, info, handler)
	case "/auth.Auth/Authenticate":
		return s.authenticate(ctx, req, info, handler)
	default:
		return handler(ctx, req)
	}
}

func (s *Service) authorize(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("unauthorized: no authorization header")
	}

	ctx, err := mid.HandleAuthorization(ctx, authHeaders[0], s.userBus, s.auth)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return handler(ctx, req)
}

func (s *Service) authenticate(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("unauthorized: no authorization header")
	}

	ctx, err := mid.HandleAuthentication(ctx, s.auth, authHeaders[0])
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	return handler(ctx, req)
}

// Token generates a token for a given key ID.
func (s *Service) Token(ctx context.Context, req *TokenRequest) (*TokenResponse, error) {
	kid := req.GetKid()
	if kid == "" {
		return nil, status.Error(codes.InvalidArgument, "kid is required")
	}

	// For simplicity, we're using a dummy claims structure here.
	// In a real implementation, you would extract claims from the context or request.
	claims := auth.Claims{
		Roles: []string{"user"},
	}

	token, err := s.auth.GenerateToken(kid, claims)
	if err != nil {
		s.log.Error(ctx, "token", "err", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	trb := TokenResponse_builder{
		Token: proto.String(token),
	}
	return trb.Build(), nil
}

// Authenticate validates a bearer token.
func (s *Service) Authenticate(ctx context.Context, req *AuthenticateRequest) (*AuthenticateResponse, error) {
	reqToken := req.GetToken()
	if reqToken == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	s.log.Info(ctx, "authenticate", "method", "grpc")

	claims, err := s.auth.Authenticate(ctx, reqToken)
	if err != nil {
		s.log.Error(ctx, "authenticate", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	arb := &AuthenticateResponse_builder{
		UserId:  proto.String(claims.Subject),
		Subject: proto.String(claims.Subject),
		Roles:   claims.Roles,
	}

	return arb.Build(), nil
}

// Authorize checks if a user is authorized for a specific rule.
func (s *Service) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
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

	err = s.auth.Authorize(ctx, claims, userID, reqRule)
	if err != nil {
		if errors.Is(err, auth.ErrForbidden) {
			return nil, status.Error(codes.PermissionDenied, "not authorized")
		}
		s.log.Error(ctx, "authorize", "err", err)
		return nil, status.Error(codes.Internal, "authorization failed")
	}

	return &AuthorizeResponse{}, nil
}

func int64ToND(in int64) *jwt.NumericDate {
	round, frac := math.Modf(float64(in))
	return jwt.NewNumericDate(time.Unix(int64(round), int64(frac*1e9)))
}
