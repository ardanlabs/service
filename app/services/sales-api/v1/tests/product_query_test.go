package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/business/core/user"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func productQuery200(t *testing.T, app appTest, sd seedData) []tableData {
	total := len(sd.admins[1].products) + len(sd.users[1].products)
	usrsMap := make(map[uuid.UUID]user.User)

	for _, adm := range sd.admins {
		usrsMap[adm.ID] = adm.User
	}

	for _, usr := range sd.users {
		usrsMap[usr.ID] = usr.User
	}

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/products?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[1].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &v1.PageDocument[productgrp.AppProductDetails]{},
			expResp: &v1.PageDocument[productgrp.AppProductDetails]{
				Page:        1,
				RowsPerPage: 10,
				Total:       total,
				Items:       toAppProductsDetails(append(sd.admins[1].products, sd.users[1].products...), usrsMap),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*v1.PageDocument[productgrp.AppProductDetails])
				exp := y.(*v1.PageDocument[productgrp.AppProductDetails])

				var found int
				for _, r := range resp.Items {
					for _, e := range exp.Items {
						if e.ID == r.ID {
							found++
							break
						}
					}
				}

				if found != total {
					return "number of expected products didn't match"
				}

				return ""
			},
		},
	}

	return table
}

func productQueryByID200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/products/%s", sd.users[1].products[0].ID),
			token:      sd.users[1].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &productgrp.AppProduct{},
			expResp:    toAppProductPtr(sd.users[1].products[0]),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
