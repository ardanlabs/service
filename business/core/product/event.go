package product

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ardanlabs/service/business/core/event"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/foundation/web"
)

func (c *Core) registerEventHandlers(evnCore *event.Core) {
	evnCore.AddHandler(user.EventSource, user.EventUpdated, c.handleUserUpdatedEvent)
}

func (c *Core) handleUserUpdatedEvent(ctx context.Context, ev event.Event) error {
	var params user.EventParamsUpdated
	err := json.Unmarshal(ev.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	c.log.Infow("user update event", "trace_id", web.GetTraceID(ctx), "user_id", params.UserID, "enabled", params.Enabled)

	// Now we can see if this user has been disabled. If they have been, we will
	// want to disable to mark all these products as deleted. Right now we don't
	// have support for this, but you can see how we can process the event.

	return nil
}
