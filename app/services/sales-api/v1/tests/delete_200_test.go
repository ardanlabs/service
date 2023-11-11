package tests

import (
	"fmt"
	"net/http"
	"testing"
)

func userDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[1].ID),
			token:      sd.users[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/users/%s", sd.admins[1].ID),
			token:      sd.admins[1].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func productDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[0].products[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/products/%s", sd.admins[0].products[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}

func homeDelete200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "asuser",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "asadmin",
			url:        fmt.Sprintf("/v1/homes/%s", sd.admins[0].homes[0].ID),
			token:      sd.admins[0].token,
			method:     http.MethodDelete,
			statusCode: http.StatusNoContent,
		},
	}

	return table
}
