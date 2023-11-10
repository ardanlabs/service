package tests

import (
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func testQuery200(t *testing.T, app appTest, sd seedData) []tableData {
	usrs := make([]user.User, 0, len(sd.admins)+len(sd.users))
	usrsMap := make(map[uuid.UUID]user.User)
	for _, adm := range sd.admins {
		usrsMap[adm.ID] = adm.User
		usrs = append(usrs, adm.User)
	}
	for _, usr := range sd.users {
		usrsMap[usr.ID] = usr.User
		usrs = append(usrs, usr.User)
	}

	table := []tableData{
		{
			name:       "user",
			url:        "/v1/users?page=1&rows=2&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[usergrp.AppUser]{},
			expResp: &response.PageDocument[usergrp.AppUser]{
				Page:        1,
				RowsPerPage: 2,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "product",
			url:        "/v1/products?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[productgrp.AppProductDetails]{},
			expResp: &response.PageDocument[productgrp.AppProductDetails]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.admins[0].products) + len(sd.users[0].products),
				Items:       toAppProductsDetails(append(sd.admins[0].products, sd.users[0].products...), usrsMap),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
		{
			name:       "home",
			url:        "/v1/homes?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &response.PageDocument[homegrp.AppHome]{},
			expResp: &response.PageDocument[homegrp.AppHome]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.admins[0].homes) + len(sd.users[0].homes),
				Items:       toAppHomes(append(sd.admins[0].homes, sd.users[0].homes...)),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
