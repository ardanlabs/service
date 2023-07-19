// Package event provides business access to events in the system.
package event

import (
	"context"

	"github.com/ardanlabs/service/foundation/logger"
)

// Core manages the set of APIs for event access.
type Core struct {
	log      *logger.Logger
	handlers map[string]map[string][]HandleFunc
}

// NewCore constructs a core for event api access.
func NewCore(log *logger.Logger) *Core {
	return &Core{
		log:      log,
		handlers: map[string]map[string][]HandleFunc{},
	}
}

// SendEvent sends event to all handlers registered for the specified event.
func (c *Core) SendEvent(ctx context.Context, event Event) error {
	c.log.Info(ctx, "sendevent", "status", "started", "source", event.Source, "type", event.Type, "params", event.RawParams)
	defer c.log.Info(ctx, "sendevent", "status", "completed")

	if m, ok := c.handlers[event.Source]; ok {
		if hfs, ok := m[event.Type]; ok {
			for _, hf := range hfs {
				c.log.Info(ctx, "sendevent", "status", "sending")

				if err := hf(ctx, event); err != nil {
					c.log.Error(ctx, "sendevent", "msg", err)
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
