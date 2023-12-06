package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/business/core/user"
	v1 "github.com/ardanlabs/service/business/web/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func homeQuery200(t *testing.T, app appTest, sd seedData) []tableData {
	total := len(sd.admins[0].homes) + len(sd.users[0].homes)
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
			url:        "/v1/homes?page=1&rows=10&orderBy=user_id,DESC",
			token:      sd.admins[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &v1.PageDocument[homegrp.AppHome]{},
			expResp: &v1.PageDocument[homegrp.AppHome]{
				Page:        1,
				RowsPerPage: 10,
				Total:       total,
				Items:       toAppHomes(append(sd.admins[0].homes, sd.users[0].homes...)),
			},
			cmpFunc: func(x interface{}, y interface{}) string {
				resp := x.(*v1.PageDocument[homegrp.AppHome])
				exp := y.(*v1.PageDocument[homegrp.AppHome])

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
					return "number of expected homes didn't match"
				}

				return ""
			},
		},
	}

	return table
}

func homeQueryByID200(t *testing.T, app appTest, sd seedData) []tableData {
	table := []tableData{
		{
			name:       "basic",
			url:        fmt.Sprintf("/v1/homes/%s", sd.users[0].homes[0].ID),
			token:      sd.users[0].token,
			statusCode: http.StatusOK,
			method:     http.MethodGet,
			resp:       &homegrp.AppHome{},
			expResp:    toAppHomePtr(sd.users[0].homes[0]),
			cmpFunc: func(x interface{}, y interface{}) string {
				return cmp.Diff(x, y)
			},
		},
	}

	return table
}
