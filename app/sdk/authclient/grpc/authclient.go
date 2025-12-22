// Package grpc provides support to access the auth service.
package grpc

import (
	"context"
	"fmt"

	"github.com/ardanlabs/service/app/domain/grpcauthapp"
	"github.com/ardanlabs/service/app/sdk/authclient"
	"github.com/ardanlabs/service/foundation/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

// Client represents a client that can talk to the auth service.
type Client struct {
	log      *logger.Logger
	url      string
	grpcConn *grpc.ClientConn
	grpc     grpcauthapp.AuthClient
}

// New constructs an Auth that can be used to talk with the auth service.
func New(log *logger.Logger, url string, options ...func(cln *Client)) (*Client, error) {
	grpcConn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth gRPC service: %w", err)
	}

	cln := Client{
		log:      log,
		url:      url,
		grpcConn: grpcConn,
		grpc:     grpcauthapp.NewAuthClient(grpcConn),
	}

	// Check if this is a gRPC URL (typically starts with "grpc://" or uses port 9000+)
	// or if there's a specific option to indicate gRPC
	for _, option := range options {
		option(&cln)
	}

	return &cln, nil
}

// WithGRPCConn allows you to provide a pre-existing gRPC connection for the auth client.
func WithGRPCConn(grpcConn *grpc.ClientConn) func(cln *Client) {
	return func(cln *Client) {
		cln.grpcConn = grpcConn
		cln.grpc = grpcauthapp.NewAuthClient(grpcConn)
	}
}

// Authenticate calls the auth service to authenticate the user.
func (cln *Client) Authenticate(ctx context.Context, authorization string) (authclient.AuthenticateResp, error) {
	arb := grpcauthapp.AuthenticateRequest_builder{
		Token: proto.String(authorization),
	}

	req := arb.Build()
	r, err := cln.grpc.Authenticate(ctx, req)
	if err != nil {
		return authclient.AuthenticateResp{}, err
	}

	return authenticateRespFromGRPC(r)
}

// Authorize calls the auth service to authorize the user.
func (cln *Client) Authorize(ctx context.Context, auth authclient.Authorize) error {
	req := authorizeRequestToGRPC(auth)

	if _, err := cln.grpc.Authorize(ctx, req); err != nil {
		return err
	}

	return nil
}

// Close closes the connection to the server.
func (cln *Client) Close() error {
	if cln.grpcConn != nil {
		return cln.grpcConn.Close()
	}

	return nil
}
