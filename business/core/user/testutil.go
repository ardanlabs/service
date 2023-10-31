package user

import (
	"context"
	"fmt"
	"math/rand"
	"net/mail"
)

// TestGenerateNewUsers is a helper method for testing.
func TestGenerateNewUsers(n int) []NewUser {
	newUsrs := make([]NewUser, n)

	for i := 0; i < n; i++ {
		idx := rand.Intn(10000)

		nu := NewUser{
			Name:            fmt.Sprintf("Name%d", idx),
			Email:           mail.Address{Address: fmt.Sprintf("Email%d", idx)},
			Roles:           []Role{RoleAdmin},
			Department:      fmt.Sprintf("Department%d", idx),
			Password:        fmt.Sprintf("Password%d", idx),
			PasswordConfirm: fmt.Sprintf("Password%d", idx),
		}

		newUsrs[i] = nu
	}

	return newUsrs
}

// TestGenerateSeedUsers is a helper method for testing.
func TestGenerateSeedUsers(n int, api *Core) ([]User, error) {
	newUsrs := TestGenerateNewUsers(n)

	usrs := make([]User, len(newUsrs))
	for i, nu := range newUsrs {
		usr, err := api.Create(context.Background(), nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user: idx: %d : %w", i, err)
		}

		usrs[i] = usr
	}

	return usrs, nil
}
