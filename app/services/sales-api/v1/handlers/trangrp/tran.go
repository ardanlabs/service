package trangrp

import (
	"context"

	"github.com/ardanlabs/service/business/data/transaction"
)

// executeUnderTransaction constructs a new Handlers value with the core apis
// using a store transaction that was created via middleware.
func (h *Handlers) executeUnderTransaction(ctx context.Context) (*Handlers, error) {
	if tx, ok := transaction.Get(ctx); ok {
		user, err := h.user.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		product, err := h.product.ExecuteUnderTransaction(tx)
		if err != nil {
			return nil, err
		}

		handlers := Handlers{
			user:    user,
			product: product,
		}

		return &handlers, nil
	}

	return h, nil
}
