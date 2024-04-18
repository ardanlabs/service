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

// Client represents a client that can talk to the auth service.
type Client struct {
	url     string
	logFunc func(ctx context.Context, s string)
	http    *http.Client
}

// New constructs an Auth that can be used to talk with the auth service.
func New(url string, logFunc func(ctx context.Context, s string), options ...func(cln *Client)) *Client {
	cln := Client{
		url:     url,
		logFunc: logFunc,
		http:    &defaultClient,
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

func (cln *Client) rawRequest(ctx context.Context, method string, url string, headers map[string]string, r io.Reader, v any) error {
	cln.logFunc(ctx, fmt.Sprintf("rawRequest: started: method: %s, url: %s", method, url))
	defer cln.logFunc(ctx, "rawRequest: completed")

	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		cln.logFunc(ctx, fmt.Sprintf("rawRequest: header: key: %s, value: %s", key, value))
		req.Header.Set(key, value)
	}

	resp, err := cln.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: error: %w", err)
	}
	defer resp.Body.Close()

	cln.logFunc(ctx, fmt.Sprintf("rawRequest: client do: status: %d", resp.StatusCode))

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
