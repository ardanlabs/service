package product

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// TestGenerateNewProducts is a helper method for testing.
func TestGenerateNewProducts(n int, userID uuid.UUID) []NewProduct {
	newPrds := make([]NewProduct, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		np := NewProduct{
			Name:     fmt.Sprintf("Name%d", idx),
			Cost:     float64(rand.Intn(500)),
			Quantity: rand.Intn(50),
			UserID:   userID,
		}

		newPrds[i] = np
	}

	return newPrds
}

// TestGenerateSeedProducts is a helper method for testing.
func TestGenerateSeedProducts(n int, api *Core, userID uuid.UUID) ([]Product, error) {
	newPrds := TestGenerateNewProducts(n, userID)

	prds := make([]Product, len(newPrds))
	for i, np := range newPrds {
		prd, err := api.Create(context.Background(), np)
		if err != nil {
			return nil, fmt.Errorf("seeding product: idx: %d : %w", i, err)
		}

		prds[i] = prd
	}

	return prds, nil
}
