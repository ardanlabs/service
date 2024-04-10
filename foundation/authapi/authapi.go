// Package authapi provides support to access the auth service.
package authapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
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

// AuthAPI represents a client that can talk to the auth service.
type AuthAPI struct {
	url     string
	logFunc func(ctx context.Context, s string)
	client  *http.Client
}

// New constructs an Auth that can be used to talk with the auth service.
func New(url string, logFunc func(ctx context.Context, s string), options ...func(authAPI *AuthAPI)) *AuthAPI {
	authAPI := AuthAPI{
		url:     url,
		logFunc: logFunc,
		client:  &defaultClient,
	}

	for _, option := range options {
		option(&authAPI)
	}

	return &authAPI
}

// WithClient adds a custom client for processing requests. It's recommend
// to not use the default client and provide your own.
func WithClient(client *http.Client) func(authAPI *AuthAPI) {
	return func(authAPI *AuthAPI) {
		authAPI.client = client
	}
}

// Authenticate calls the auth service to authenticate the user.
func (api *AuthAPI) Authenticate(ctx context.Context, authorization string) (AuthResp, error) {
	endpoint := fmt.Sprintf("%s/v1/auth/authenticate", api.url)

	headers := map[string]string{
		"authorization": authorization,
	}

	var resp AuthResp
	if err := api.rawRequest(ctx, http.MethodGet, endpoint, headers, nil, &resp); err != nil {
		return AuthResp{}, err
	}

	return resp, nil
}

// Authorize calls the auth service to authorize the user.
func (api *AuthAPI) Authorize(ctx context.Context, authInfo AuthInfo) error {
	endpoint := fmt.Sprintf("%s/v1/auth/authorize", api.url)

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(authInfo); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}

	if err := api.rawRequest(ctx, http.MethodPost, endpoint, nil, &b, nil); err != nil {
		return err
	}

	return nil
}

func (api *AuthAPI) rawRequest(ctx context.Context, method string, url string, headers map[string]string, r io.Reader, v any) error {
	api.logFunc(ctx, fmt.Sprintf("rawRequest: started: method: %s, url: %s", method, url))
	defer api.logFunc(ctx, "rawRequest: completed")

	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		api.logFunc(ctx, fmt.Sprintf("rawRequest: header: key: %s, value: %s", key, value))
		req.Header.Set(key, value)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("do: error: %w", err)
	}
	defer resp.Body.Close()

	api.logFunc(ctx, fmt.Sprintf("rawRequest: client do: status: %d", resp.StatusCode))

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil

	case http.StatusOK:
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return nil

	case http.StatusUnauthorized:
		var err Error
		if err := json.Unmarshal(data, &err); err != nil {
			return fmt.Errorf("failed: response: %s, decoding error: %w ", string(data), err)
		}
		return err

	default:
		return fmt.Errorf("failed: response: %s", string(data))
	}
}
