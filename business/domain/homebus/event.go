package homebus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/delegate"
)

// registerDelegateFunctions will register action functions with the delegate
// system. If the business was constructed for query only, there won't be a
// delegate provided.
func (b *Business) registerDelegateFunctions() {
	if b.delegate != nil {
		b.delegate.Register(userbus.DomainName, userbus.ActionDeleted, b.actionUserDeleted)
	}
}

// actionUserDeleted is executed by the user domain indirectly when a user is deleted.
func (b *Business) actionUserDeleted(ctx context.Context, data delegate.Data) error {
	var params userbus.ActionDeletedParms
	err := json.Unmarshal(data.RawParams, &params)
	if err != nil {
		return fmt.Errorf("expected an encoded %T: %w", params, err)
	}

	b.log.Info(ctx, "action-userdeleted", "user_id", params.UserID)

	// Now we can mark all the homes for this user as deleted.

	return nil
}
