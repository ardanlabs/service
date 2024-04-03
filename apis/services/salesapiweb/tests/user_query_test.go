package tests

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/page"
	"github.com/ardanlabs/service/business/core/crud/user"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/google/go-cmp/cmp"
)

func userQuery200(sd dbtest.SeedData) []dbtest.AppTable {
	usrs := make([]user.User, 0, len(sd.Admins)+len(sd.Users))

	for _, adm := range sd.Admins {
		usrs = append(usrs, adm.User)
	}

	for _, usr := range sd.Users {
		usrs = append(usrs, usr.User)
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].ID.String() <= usrs[j].ID.String()
	})

	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        "/v1/users?page=1&rows=10&orderBy=user_id,ASC&name=Name",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &page.Document[userapp.User]{},
			ExpResp: &page.Document[userapp.User]{
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

func userQueryByID200(sd dbtest.SeedData) []dbtest.AppTable {
	table := []dbtest.AppTable{
		{
			Name:       "basic",
			URL:        fmt.Sprintf("/v1/users/%s", sd.Users[0].ID),
			Token:      sd.Users[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			Resp:       &userapp.User{},
			ExpResp:    toAppUserPtr(sd.Users[0].User),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
