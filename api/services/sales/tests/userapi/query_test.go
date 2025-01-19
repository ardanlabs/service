package user_test

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/domain/userapp"
	"github.com/ardanlabs/service/app/sdk/apitest"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/query"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/google/go-cmp/cmp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	usrs := make([]userbus.User, 0, len(sd.Admins)+len(sd.Users))

	for _, adm := range sd.Admins {
		usrs = append(usrs, adm.User)
	}

	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/users?page=1&rows=10&orderBy=user_id,ASC&name=Name",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[userapp.User]{},
			ExpResp: &query.Result[userapp.User]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(usrs),
				Items:       toAppUsers(usrs),
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func query400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/v1/users?page=1&rows=10&email=a.com",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"email\",\"error\":\"mail: missing '@' or angle-addr\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-orderby-value",
			URL:        "/v1/users?page=1&rows=10&orderBy=ser_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Newf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}

func queryByID200(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &userapp.User{},
			ExpResp:    toAppUserPtr(sd.Users[0].User),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
