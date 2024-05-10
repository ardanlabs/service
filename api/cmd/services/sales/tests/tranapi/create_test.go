package tran_test

import (
	"net/http"

	"github.com/ardanlabs/service/api/sdk/http/apitest"
	"github.com/ardanlabs/service/app/domain/tranapp"
	"github.com/google/go-cmp/cmp"
)

func create200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/tranexample",
			Token:      sd.Admins[0].Token,
			Method:     http.MethodPost,
			StatusCode: http.StatusOK,
			Input: &tranapp.NewTran{
				Product: tranapp.NewProduct{
					Name:     "Guitar",
					Cost:     10.34,
					Quantity: 10,
				},
				User: tranapp.NewUser{
					Name:            "Bill Kennedy",
					Email:           "bill@ardanlabs.com",
					Roles:           []string{"ADMIN"},
					Department:      "IT",
					Password:        "123",
					PasswordConfirm: "123",
				},
			},
			GotResp: &tranapp.Product{},
			ExpResp: &tranapp.Product{
				Name:     "Guitar",
				Cost:     10.34,
				Quantity: 10,
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*tranapp.Product)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*tranapp.Product)

				expResp.ID = gotResp.ID
				expResp.UserID = gotResp.UserID
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
