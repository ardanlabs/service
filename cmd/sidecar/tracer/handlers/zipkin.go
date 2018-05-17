package handlers

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ardanlabs/service/internal/platform/web"
	"github.com/openzipkin/zipkin-go/model"
	"go.opencensus.io/trace"
)

// Zipkin represents the API to collect span data and send to zipkin.
type Zipkin struct {
	zipkinHost  string        // IP:port of the zipkin service.
	localHost   string        // IP:port of the sidecare consuming the trace data.
	sendTimeout time.Duration // Time to wait for the sidecar to respond on send.
	client      http.Client   // Provides APIs for performing the http send.
}

// NewZipkin provides support for publishing traces to zipkin.
func NewZipkin(zipkinHost string, localHost string, sendTimeout time.Duration) *Zipkin {
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	z := Zipkin{
		zipkinHost:  zipkinHost,
		localHost:   localHost,
		sendTimeout: sendTimeout,
		client: http.Client{
			Transport: &tr,
		},
	}

	return &z
}

// Publish takes a batch and publishes that to a host system.
func (z *Zipkin) Publish(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var sd []trace.SpanData
	if err := json.NewDecoder(r.Body).Decode(&sd); err != nil {
		return err
	}

	if err := z.send(sd); err != nil {
		return err
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)

	return nil
}

// send uses HTTP to send the data to the tracing sidecare for processing.
func (z *Zipkin) send(sendBatch []trace.SpanData) error {
	le, err := newEndpoint("crud", z.localHost)
	if err != nil {
		return err
	}

	sm := convertForZipkin(sendBatch, le)
	data, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", z.zipkinHost, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(req.Context(), z.sendTimeout)
	defer cancel()
	req = req.WithContext(ctx)

	ch := make(chan error)
	go func() {
		resp, err := z.client.Do(req)
		if err != nil {
			ch <- err
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusAccepted {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- fmt.Errorf("error on call : status[%s]", resp.Status)
				return
			}
			ch <- fmt.Errorf("error on call : status[%s] : %s", resp.Status, string(data))
			return
		}

		ch <- nil
	}()

	return <-ch
}

// =============================================================================

const (
	statusCodeTagKey        = "error"
	statusDescriptionTagKey = "opencensus.status_description"
)

var (
	sampledTrue    = true
	canonicalCodes = [...]string{
		"OK",
		"CANCELLED",
		"UNKNOWN",
		"INVALID_ARGUMENT",
		"DEADLINE_EXCEEDED",
		"NOT_FOUND",
		"ALREADY_EXISTS",
		"PERMISSION_DENIED",
		"RESOURCE_EXHAUSTED",
		"FAILED_PRECONDITION",
		"ABORTED",
		"OUT_OF_RANGE",
		"UNIMPLEMENTED",
		"INTERNAL",
		"UNAVAILABLE",
		"DATA_LOSS",
		"UNAUTHENTICATED",
	}
)

func convertForZipkin(spanData []trace.SpanData, localEndpoint *model.Endpoint) []model.SpanModel {
	sm := make([]model.SpanModel, len(spanData))
	for i := range spanData {
		sm[i] = zipkinSpan(&spanData[i], localEndpoint)
	}
	return sm
}

func newEndpoint(serviceName string, hostPort string) (*model.Endpoint, error) {
	e := &model.Endpoint{
		ServiceName: serviceName,
	}

	if hostPort == "" || hostPort == ":0" {
		if serviceName == "" {
			// if all properties are empty we should not have an Endpoint object.
			return nil, nil
		}
		return e, nil
	}

	if strings.IndexByte(hostPort, ':') < 0 {
		hostPort += ":0"
	}

	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}

	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, err
	}
	e.Port = uint16(p)

	addrs, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	for i := range addrs {
		addr := addrs[i].To4()
		if addr == nil {
			// IPv6 - 16 bytes
			if e.IPv6 == nil {
				e.IPv6 = addrs[i].To16()
			}
		} else {
			// IPv4 - 4 bytes
			if e.IPv4 == nil {
				e.IPv4 = addr
			}
		}
		if e.IPv4 != nil && e.IPv6 != nil {
			// Both IPv4 & IPv6 have been set, done...
			break
		}
	}

	// default to 0 filled 4 byte array for IPv4 if IPv6 only host was found
	if e.IPv4 == nil {
		e.IPv4 = make([]byte, 4)
	}

	return e, nil
}

func canonicalCodeString(code int32) string {
	if code < 0 || int(code) >= len(canonicalCodes) {
		return "error code " + strconv.FormatInt(int64(code), 10)
	}
	return canonicalCodes[code]
}

func convertTraceID(t trace.TraceID) model.TraceID {
	return model.TraceID{
		High: binary.BigEndian.Uint64(t[:8]),
		Low:  binary.BigEndian.Uint64(t[8:]),
	}
}

func convertSpanID(s trace.SpanID) model.ID {
	return model.ID(binary.BigEndian.Uint64(s[:]))
}

func spanKind(s *trace.SpanData) model.Kind {
	switch s.SpanKind {
	case trace.SpanKindClient:
		return model.Client
	case trace.SpanKindServer:
		return model.Server
	}
	return model.Undetermined
}

func zipkinSpan(s *trace.SpanData, localEndpoint *model.Endpoint) model.SpanModel {
	sc := s.SpanContext
	z := model.SpanModel{
		SpanContext: model.SpanContext{
			TraceID: convertTraceID(sc.TraceID),
			ID:      convertSpanID(sc.SpanID),
			Sampled: &sampledTrue,
		},
		Kind:          spanKind(s),
		Name:          s.Name,
		Timestamp:     s.StartTime,
		Shared:        false,
		LocalEndpoint: localEndpoint,
	}

	if s.ParentSpanID != (trace.SpanID{}) {
		id := convertSpanID(s.ParentSpanID)
		z.ParentID = &id
	}

	if s, e := s.StartTime, s.EndTime; !s.IsZero() && !e.IsZero() {
		z.Duration = e.Sub(s)
	}

	// construct Tags from s.Attributes and s.Status.
	if len(s.Attributes) != 0 {
		m := make(map[string]string, len(s.Attributes)+2)
		for key, value := range s.Attributes {
			switch v := value.(type) {
			case string:
				m[key] = v
			case bool:
				if v {
					m[key] = "true"
				} else {
					m[key] = "false"
				}
			case int64:
				m[key] = strconv.FormatInt(v, 10)
			}
		}
		z.Tags = m
	}
	if s.Status.Code != 0 || s.Status.Message != "" {
		if z.Tags == nil {
			z.Tags = make(map[string]string, 2)
		}
		if s.Status.Code != 0 {
			z.Tags[statusCodeTagKey] = canonicalCodeString(s.Status.Code)
		}
		if s.Status.Message != "" {
			z.Tags[statusDescriptionTagKey] = s.Status.Message
		}
	}

	// construct Annotations from s.Annotations and s.MessageEvents.
	if len(s.Annotations) != 0 || len(s.MessageEvents) != 0 {
		z.Annotations = make([]model.Annotation, 0, len(s.Annotations)+len(s.MessageEvents))
		for _, a := range s.Annotations {
			z.Annotations = append(z.Annotations, model.Annotation{
				Timestamp: a.Time,
				Value:     a.Message,
			})
		}
		for _, m := range s.MessageEvents {
			a := model.Annotation{
				Timestamp: m.Time,
			}
			switch m.EventType {
			case trace.MessageEventTypeSent:
				a.Value = "SENT"
			case trace.MessageEventTypeRecv:
				a.Value = "RECV"
			default:
				a.Value = "<?>"
			}
			z.Annotations = append(z.Annotations, a)
		}
	}

	return z
}
