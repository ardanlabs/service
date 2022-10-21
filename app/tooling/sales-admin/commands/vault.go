package commands

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

func Vault() error {
	c, err := api.NewClient(&api.Config{
		Address: "http://0.0.0.0:8200",
	})
	if err != nil {
		return fmt.Errorf("new: %w", err)
	}

	v2 := c.KVv2("secret")

	if _, err := v2.Put(context.Background(), "service", map[string]interface{}{"pk": "12233"}); err != nil {
		return fmt.Errorf("put: %w", err)
	}

	sec, err := v2.Get(context.Background(), "service")
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	fmt.Println(sec.Data)

	return nil
}
