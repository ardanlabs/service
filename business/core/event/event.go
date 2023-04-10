// Package event provides business access to events in the system.
package event

import (
	"context"

	"github.com/ardanlabs/service/foundation/web"
	"go.uber.org/zap"
)

// Core manages the set of APIs for event access.
type Core struct {
	log      *zap.SugaredLogger
	handlers map[string]map[string][]HandleFunc
}

// NewCore constructs a core for event api access.
func NewCore(log *zap.SugaredLogger) *Core {
	return &Core{
		log:      log,
		handlers: map[string]map[string][]HandleFunc{},
	}
}

// SendEvent sends event to all handlers registered for the specified event.
func (c *Core) SendEvent(ctx context.Context, event Event) error {
	c.log.Infow("sendevent", "trace_id", web.GetTraceID(ctx), "status", "started", "source", event.Source, "type", event.Type, "params", event.RawParams)
	defer c.log.Infow("sendevent", "trace_id", web.GetTraceID(ctx), "status", "completed")

	if m, ok := c.handlers[event.Source]; ok {
		if hfs, ok := m[event.Type]; ok {
			for _, hf := range hfs {
				c.log.Infow("sendevent", "trace_id", web.GetTraceID(ctx), "status", "sending")

				if err := hf(ctx, event); err != nil {
					c.log.Infof("sendevent", "trace_id", web.GetTraceID(ctx), "ERROR", err)
				}
			}
		}
	}

	return nil
}

// AddHandler add handler to specific event from specific source.
func (c *Core) AddHandler(source, t string, f HandleFunc) {
	ss, ok := c.handlers[source]
	if !ok {
		ss = map[string][]HandleFunc{}
	}

	ss[t] = append(ss[t], f)
	c.handlers[source] = ss
}
