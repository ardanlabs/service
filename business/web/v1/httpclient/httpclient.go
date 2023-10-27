// Package httpclient implements a wrapper of http client supports logging, tracing, and proxy
package httpclient

import (
	"context"
	"fmt"
	"github.com/ardanlabs/service/foundation/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"time"
)

type options struct {
	logger  *logger.Logger
	logBody bool
	tracing bool
	proxy   func(request *http.Request) (*url.URL, error)
}

type Option func(o *options)

func WithLogger(l *logger.Logger, body bool) Option {
	return func(o *options) {
		o.logger = l
		o.logBody = body
	}
}

func WithTracing() Option {
	return func(o *options) {
		o.tracing = true
	}
}

func WithProxy(proxy *url.URL) Option {
	return func(o *options) {
		if proxy != nil {
			o.proxy = http.ProxyURL(proxy)
		}
	}
}

type roundTripperFn func(req *http.Request) (*http.Response, error)

func (f roundTripperFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var roundTripper http.RoundTripper = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          1000,
	MaxIdleConnsPerHost:   100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// New returns a HTTP client with logging, tracing, and proxy support
func New(opts ...Option) *http.Client {
	o := new(options)
	for _, opt := range opts {
		opt(o)
	}

	rt := roundTripper
	if t, ok := rt.(*http.Transport); ok && o.proxy != nil {
		t = t.Clone()
		t.Proxy = o.proxy
		rt = t
	}

	if o.tracing {
		rt = otelhttp.NewTransport(rt, otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}))
	}

	if o.logger != nil {
		rt = logRoundTripper(rt, o.logger, o.logBody)
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: rt,
	}
}

func logRoundTripper(rt http.RoundTripper, l *logger.Logger, body bool) http.RoundTripper {
	return roundTripperFn(func(req *http.Request) (resp *http.Response, err error) {
		ctx := req.Context()
		start := time.Now()

		args := []any{
			slog.String("http.client.host", req.URL.Host),
			slog.String("http.client.path", req.URL.Path),
		}

		if body {
			b, err := httputil.DumpRequest(req, true)
			if err != nil {
				return nil, fmt.Errorf("dump http request: %w", err)
			}
			argsWithBody := append(args, slog.Any("http.client.request", string(b)))
			l.Info(ctx, "http.client: sending request", argsWithBody...)
		} else {
			l.Info(ctx, "http.client: sending request", args...)
		}

		defer func() {
			args = append(args, slog.String("http.client.latency", time.Since(start).String()))
			if err != nil {
				args = append(args, slog.Any("error", err))
			}
			l.Info(ctx, "http.client: received response", args...)
		}()

		resp, err = rt.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		args = append(args, slog.Int("http.client.status", resp.StatusCode))
		if body {
			b, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return resp, fmt.Errorf("dump http response: %w", err)
			}
			args = append(args, slog.String("http.client.response", string(b)))
		}
		return resp, nil
	})
}
