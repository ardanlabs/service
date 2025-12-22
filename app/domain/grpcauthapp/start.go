package grpcauthapp

import (
	"context"
	"net"

	"github.com/ardanlabs/service/app/sdk/auth"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/foundation/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Config holds the dependencies for the gRPC service.
type Config struct {
	Log      *logger.Logger
	Auth     *auth.Auth
	Listener net.Listener
	UserBus  userbus.ExtBusiness
}

// Start constructs the registers the auth app to the grpc server.
func Start(ctx context.Context, cfg Config) *App {
	cfg.Log.Info(context.Background(), "auth service", "status", "start auth server")

	api := newApp(cfg)

	gs := grpc.NewServer(
		grpc.UnaryInterceptor(api.authInterceptor),
	)

	api.gs = gs

	RegisterAuthServer(gs, api)
	reflection.Register(gs)

	go func() {
		if err := gs.Serve(cfg.Listener); err != nil {
			api.log.Error(ctx, "startup", "status", "auth server error", "err", err)
		}
	}()

	return api
}
