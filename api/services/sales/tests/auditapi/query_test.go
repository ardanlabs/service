package audit_test

import (
	"net/http"
	"sort"
	"strings"

	"github.com/ardanlabs/service/app/domain/auditapp"
	"github.com/ardanlabs/service/app/sdk/apitest"
	"github.com/ardanlabs/service/app/sdk/errs"
	"github.com/ardanlabs/service/app/sdk/query"
	"github.com/google/go-cmp/cmp"
)

func query200(sd apitest.SeedData) []apitest.Table {
	sort.Slice(sd.Admins[0].Audits, func(i, j int) bool {
		return sd.Admins[0].Audits[i].ObjName.String() <= sd.Admins[0].Audits[j].ObjName.String()
	})

	table := []apitest.Table{
		{
			Name:       "basic",
			URL:        "/v1/audits?page=1&rows=10&orderBy=obj_name,ASC&obj_name=ObjName",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusOK,
			Method:     http.MethodGet,
			GotResp:    &query.Result[auditapp.Audit]{},
			ExpResp: &query.Result[auditapp.Audit]{
				Page:        1,
				RowsPerPage: 10,
				Total:       len(sd.Admins[0].Audits),
				Items:       toAppAudits(sd.Admins[0].Audits),
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(*query.Result[auditapp.Audit])
				if !exists {
					return "error occurred"
				}

				expResp := exp.(*query.Result[auditapp.Audit])

				for i := range gotResp.Items {
					if gotResp.Items[i].Timestamp == expResp.Items[i].Timestamp {
						expResp.Items[i].Timestamp = gotResp.Items[i].Timestamp
					}

					gotResp.Items[i].Data = strings.ReplaceAll(gotResp.Items[i].Data, " ", "")
					expResp.Items[i].Data = strings.ReplaceAll(expResp.Items[i].Data, " ", "")
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func query400(sd apitest.SeedData) []apitest.Table {
	table := []apitest.Table{
		{
			Name:       "bad-query-filter",
			URL:        "/v1/audits?page=1&rows=10&obj_id=123",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"obj_id\",\"error\":\"invalid UUID length: 3\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:       "bad-orderby-value",
			URL:        "/v1/audits?page=1&rows=10&orderBy=ser_id,ASC",
			Token:      sd.Admins[0].Token,
			StatusCode: http.StatusBadRequest,
			Method:     http.MethodGet,
			GotResp:    &errs.Error{},
			ExpResp:    errs.Errorf(errs.InvalidArgument, "[{\"field\":\"order\",\"error\":\"unknown order: ser_id\"}]"),
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
