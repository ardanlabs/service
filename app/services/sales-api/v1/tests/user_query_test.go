package tests

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/business/core/crud/user"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func userQuery200(sd seedData) []tableData {
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
			name:       "basic",
			url:        "/v1/users?page=1&rows=10&orderBy=user_id,ASC&name=Name",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &v1.PageDocument[usergrp.AppUser]{},
			expResp: &v1.PageDocument[usergrp.AppUser]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*v1.PageDocument[usergrp.AppUser])
				exp := y.(*v1.PageDocument[usergrp.AppUser])

				var found int
				for _, r := range resp.Items {
					for _, e := range exp.Items {
						if e.ID == r.ID {
							found++
							break
						}
					}
				}

				if found != len(usrs) {
					return "number of expected users didn't match"
				}

				return ""
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
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
