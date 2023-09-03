package home

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

// TestGenerateNewHomes is a helper method for testing.
func TestGenerateNewHomes(n int, userID uuid.UUID) []NewHome {
	newHmes := make([]NewHome, n)

	idx := rand.Intn(10000)
	zipCode := fmt.Sprintf("%05d", idx)

	for i := 0; i < n; i++ {
		nh := NewHome{
			Type: fmt.Sprintf("Type%d", idx),
			Address: Address{
				Address1: fmt.Sprintf("Address%d", idx),
				Address2: fmt.Sprintf("Address%d", idx),
				ZipCode:  zipCode,
				City:     fmt.Sprintf("City%d", idx),
				State:    fmt.Sprintf("State%d", idx),
				Country:  fmt.Sprintf("Country%d", idx),
			},
			UserID: userID,
		}

		newHmes[i] = nh
		idx++
	}

	return newHmes
}

// TestGenerateSeedHomes is a helper method for testing.
func TestGenerateSeedHomes(n int, api *Core, userID uuid.UUID) ([]Home, error) {
	newHmes := TestGenerateNewHomes(n, userID)

	hmes := make([]Home, len(newHmes))
	for i, np := range newHmes {
		prd, err := api.Create(context.Background(), np)
		if err != nil {
			return nil, fmt.Errorf("seeding home: idx: %d : %w", i, err)
		}

		hmes[i] = prd
	}

	return hmes, nil
}

// ParseAddress is a helper function to create an address value.
func ParseAddress(address1 string, address2 string, zipCode string, city string, state string, country string) Address {
	return Address{
		Address1: address1,
		Address2: address2,
		ZipCode:  zipCode,
		City:     city,
		State:    state,
		Country:  country,
	}
}
