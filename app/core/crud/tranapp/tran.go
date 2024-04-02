package tranapp

import (
	"context"

	"github.com/ardanlabs/service/business/data/transaction"
)

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (c *Core) executeUnderTransaction(ctx context.Context) (*Core, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := c.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		product, err := c.product.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		handlers := Core{
			user:    user,
			product: product,
		}

		return &handlers, nil
	}

	return c, nil
}
