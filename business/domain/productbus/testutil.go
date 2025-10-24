package productbus

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
	"github.com/google/uuid"
)

// TestGenerateNewProducts is a helper method for testing.
func TestGenerateNewProducts(n int, userID uuid.UUID) []NewProduct {
	newPrds := make([]NewProduct, n)

	idx := rand.Intn(10000)
	for i := range n {
		idx++

		np := NewProduct{
			Name:     name.MustParse(fmt.Sprintf("Name%d", idx)),
			Cost:     money.MustParse(float64(rand.Intn(500) + 1)),
			Quantity: quantity.MustParse(rand.Intn(50) + 1),
			UserID:   userID,
		}

		newPrds[i] = np
	}

	return newPrds
}

// TestGenerateSeedProducts is a helper method for testing.
func TestGenerateSeedProducts(ctx context.Context, n int, api *Business, userID uuid.UUID) ([]Product, error) {
	newPrds := TestGenerateNewProducts(n, userID)

	prds := make([]Product, len(newPrds))
	for i, np := range newPrds {
		prd, err := api.Create(ctx, np)
		if err != nil {
			return nil, fmt.Errorf("seeding product: idx: %d : %w", i, err)
		}

		prds[i] = prd
	}

	return prds, nil
}
