// Package authclient provides support to access the auth service.
package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/tracer"
	"go.opentelemetry.io/otel/attribute"
)

// This provides a default client configuration, but it's recommended
// this is replaced by the user with application specific settings using
// the WithClient function at the time a AuthAPI is constructed.
var defaultClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// Client represents a client that can talk to the auth service.
type Client struct {
	log  *logger.Logger
	url  string
	http *http.Client
}

// New constructs an Auth that can be used to talk with the auth service.
func New(log *logger.Logger, url string, options ...func(cln *Client)) *Client {
	cln := Client{
		log:  log,
		url:  url,
		http: &defaultClient,
	}

	for _, option := range options {
		option(&cln)
	}

	return &cln
}

// WithClient adds a custom client for processing requests. It's recommend
// to not use the default client and provide your own.
func WithClient(http *http.Client) func(cln *Client) {
	return func(cln *Client) {
		cln.http = http
	}
}

// Authenticate calls the auth service to authenticate the user.
func (cln *Client) Authenticate(ctx context.Context, authorization string) (AuthenticateResp, error) {
	endpoint := fmt.Sprintf("%s/v1/auth/authenticate", cln.url)

	headers := map[string]string{
		"authorization": authorization,
	}

	var resp AuthenticateResp
	if err := cln.rawRequest(ctx, http.MethodGet, endpoint, headers, nil, &resp); err != nil {
		return AuthenticateResp{}, err
	}

	return resp, nil
}

// Authorize calls the auth service to authorize the user.
func (cln *Client) Authorize(ctx context.Context, auth Authorize) error {
	endpoint := fmt.Sprintf("%s/v1/auth/authorize", cln.url)

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(auth); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	if err := cln.rawRequest(ctx, http.MethodPost, endpoint, nil, &b, nil); err != nil {
		return err
	}

	return nil
}

func (cln *Client) rawRequest(ctx context.Context, method string, endpoint string, headers map[string]string, r io.Reader, v any) error {
	var statusCode int

	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint: %w", err)
	}
	base := path.Base(u.Path)

	cln.log.Info(ctx, "authclient: rawRequest: started", "method", method, "call", base, "endpoint", endpoint)
	defer func() {
		cln.log.Info(ctx, "authclient: rawRequest: completed", "status", statusCode)
	}()

	ctx, span := tracer.AddSpan(ctx, fmt.Sprintf("app.api.authclient.%s", base), attribute.String("endpoint", endpoint))
	defer func() {
		span.SetAttributes(attribute.Int("status", statusCode))
		span.End()
	}()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		cln.log.Info(ctx, "authclient: rawRequest", "key", key, "value", value)
		req.Header.Set(key, value)
	}

	resp, err := cln.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: error: %w", err)
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode

	if statusCode == http.StatusNoContent {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	switch statusCode {
	case http.StatusNoContent:
		return nil

	case http.StatusOK:
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return nil

	case http.StatusUnauthorized:
		var err *errs.Error
		if err := json.Unmarshal(data, &err); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return err

	default:
		return fmt.Errorf("failed: response: %s", string(data))
	}
}
