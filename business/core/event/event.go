// Package event provides business access to events in the system.
package event

import (
	"context"

	"github.com/ardanlabs/service/foundation/logger"
)

// Core manages the set of APIs for event access.
type Core struct {
	log   *logger.Logger
	funcs map[string]map[string][]Func
}

// NewCore constructs a core for event api access.
func NewCore(log *logger.Logger) *Core {
	return &Core{
		log:   log,
		funcs: map[string]map[string][]Func{},
	}
}

// Register adds an event function to an event source and type.
func (c *Core) Register(source string, typ string, f Func) {
	ss, ok := c.funcs[source]
	if !ok {
		ss = map[string][]Func{}
	}

	ss[typ] = append(ss[typ], f)
	c.funcs[source] = ss
}

// Execute executes all event functions registered for the specified event.
// These functions are executed synchronously on the G making the call.
func (c *Core) Execute(ctx context.Context, event Event) error {
	c.log.Info(ctx, "event execute", "status", "started", "source", event.Source, "type", event.Type, "params", event.RawParams)
	defer c.log.Info(ctx, "event execute", "status", "completed")

	if m, ok := c.funcs[event.Source]; ok {
		if hfs, ok := m[event.Type]; ok {
			for _, hf := range hfs {
				c.log.Info(ctx, "event execute", "status", "sending")

				if err := hf(ctx, event); err != nil {
					c.log.Error(ctx, "event execute", "msg", err)
				}
			}
		}
	}

	return nil
}
