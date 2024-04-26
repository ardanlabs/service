package tranapp

import (
	"context"

	"github.com/ardanlabs/service/business/api/transaction"
)

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (c *Core) executeUnderTransaction(ctx context.Context) (*Core, error) {
	if tx, ok := transaction.Get(ctx); ok {
		userBus, err := c.userBus.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		productBus, err := c.productBus.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		handlers := Core{
			userBus:    userBus,
			productBus: productBus,
		}

		return &handlers, nil
	}

	return c, nil
}
