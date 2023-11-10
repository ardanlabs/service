package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func testDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "product-user",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "product-admin",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-user",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "home-admin",
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-user",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
			token:      sd.users[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "user-admin",
			url:        fmt.Sprintf("/v1/users/%s", sd.admins[1].ID),
			token:      sd.admins[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}
