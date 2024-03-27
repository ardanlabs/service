package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/services/sales-api/handlers/crud/usergrp"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/web/page"
	"github.com/google/go-cmp/cmp"
)

func userQuery200(sd seedData) []tableData {
	usrs := make([]user.User, 0, len(sd.admins)+len(sd.users))

	for _, adm := range sd.admins {
		usrs = append(usrs, adm.User)
	}

	for _, usr := range sd.users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []tableData{
		{
			name:       "basic",
			url:        "/v1/users?page=1&rows=10&orderBy=user_id,ASC&name=Name",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &page.Document[usergrp.AppUser]{},
			expResp: &page.Document[usergrp.AppUser]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func userQueryByID200(sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/users/%s", sd.users[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &usergrp.AppUser{},
			expResp:    toAppUserPtr(sd.users[0].User),
			cmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
