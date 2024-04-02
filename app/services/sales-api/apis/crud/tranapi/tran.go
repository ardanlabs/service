package tranapi

import (
	"context"

	"github.com/ardanlabs/service/business/data/transaction"
)

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (api *API) executeUnderTransaction(ctx context.Context) (*API, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := api.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		product, err := api.product.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		handlers := API{
			user:    user,
			product: product,
		}

		return &handlers, nil
	}

	return api, nil
}
