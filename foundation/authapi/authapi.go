// Package authapi provides support to access the auth service.
package authapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// This provides a default client configuration, but it's recommended
// this is replaced by the user with application specific settings using
// the WithClient function at the time a GraphQL is constructed.
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

// Auth represents a client that can talk to the auth service.
type Auth struct {
	url     string
	client  *http.Client
	logFunc func(s string)
}

// New constructs an Auth that can be used to talk with the auth service.
func New(url string, options ...func(auth *Auth)) *Auth {
	auth := Auth{
		url:    url,
		client: &defaultClient,
	}

	for _, option := range options {
		option(&auth)
	}

	return &auth
}

// WithClient adds a custom client for processing requests. It's recommend
// to not use the default client and provide your own.
func WithClient(client *http.Client) func(auth *Auth) {
	return func(auth *Auth) {
		auth.client = client
	}
}

// WithLogging acceps a function for capturing raw execution messages for the
// purpose of application logging.
func WithLogging(logFunc func(s string)) func(auth *Auth) {
	return func(auth *Auth) {
		auth.logFunc = logFunc
	}
}

// Token returns a token for the specified user and kid.
func (ath *Auth) Token(ctx context.Context, userName string, password string, kid string) (Token, error) {
	endpoint := fmt.Sprintf("%s/users/token/%s", ath.url, kid)

	userPass := fmt.Sprintf("%s:%s", userName, password)
	headers := map[string]string{
		"Basic": base64.StdEncoding.EncodeToString([]byte(userPass)),
	}

	var token Token
	if err := ath.rawRequest(ctx, http.MethodGet, endpoint, headers, nil, &token); err != nil {
		return Token{}, err
	}

	return token, nil
}

func (ath *Auth) rawRequest(ctx context.Context, method string, url string, headers map[string]string, r io.Reader, response interface{}) error {

	// Use the TeeReader to capture the request being sent. This is needed if the
	// requrest fails for the error being returned or for logging if a log
	// function is provided. The TeeReader will write the request to this buffer
	// during the http operation.
	var request bytes.Buffer
	r = io.TeeReader(r, &request)

	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := ath.client.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("op error: status code: %s", resp.Status)
	}

	if ath.logFunc != nil {
		ath.logFunc(fmt.Sprintf("request:[%s] data:[%s]", request.String(), string(data)))
	}

	result := struct {
		Data   interface{}
		Errors []struct {
			Message string
		}
	}{
		Data: response,
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("decoding error: %w response: %s", err, string(data))
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("op error: request:[%s] error:[%s]", request.String(), result.Errors[0].Message)
	}

	return nil
}
