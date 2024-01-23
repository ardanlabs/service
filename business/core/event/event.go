// Package event provides business access to events in the system.
package event

import (
	"context"

	"github.com/ardanlabs/service/foundation/logger"
)

// Core manages the set of APIs for event access.
type Core struct {
	log      *logger.Logger
	eventFns map[string]map[string][]EventFn
}

// NewCore constructs a core for event api access.
func NewCore(log *logger.Logger) *Core {
	return &Core{
		log:      log,
		eventFns: map[string]map[string][]EventFn{},
	}
}

// Register adds an event function to an event source and type.
func (c *Core) Register(source string, typ string, eventFn EventFn) {
	sourceMap, ok := c.eventFns[source]
	if !ok {
		sourceMap = map[string][]EventFn{}
	}

	sourceMap[typ] = append(sourceMap[typ], eventFn)
	c.eventFns[source] = sourceMap
}

// Execute executes all event functions registered for the specified event.
// These functions are executed synchronously on the G making the call.
func (c *Core) Execute(ctx context.Context, event Event) error {
	c.log.Info(ctx, "event execute", "status", "started", "source", event.Source, "type", event.Type, "params", event.RawParams)
	defer c.log.Info(ctx, "event execute", "status", "completed")

	if sourceMap, ok := c.eventFns[event.Source]; ok {
		if eventFns, ok := sourceMap[event.Type]; ok {
			for _, eventFn := range eventFns {
				c.log.Info(ctx, "event execute", "status", "sending")

				if err := eventFn(ctx, event); err != nil {
					c.log.Error(ctx, "event execute", "msg", err)
				}
			}
		}
	}

	return nil
}
