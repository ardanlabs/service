package tranapp

import (
	"context"

	"github.com/ardanlabs/service/business/api/transaction"
)

// executeUnderTransaction constructs a new Handlers value with the domain apis
// using a store transaction that was created via middleware.
func (a *App) executeUnderTransaction(ctx context.Context) (*App, error) {
	if tx, ok := transaction.Get(ctx); ok {
		userBus, err := a.userBus.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		productBus, err := a.productBus.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		handlers := App{
			userBus:    userBus,
			productBus: productBus,
		}

		return &handlers, nil
	}

	return a, nil
}
