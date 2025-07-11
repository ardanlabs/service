package auditbus_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/dbtest"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/unittest"
	"github.com/ardanlabs/service/business/types/domain"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/google/go-cmp/cmp"
)

func Test_Audit(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Audit")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unittest.Run(t, query(db.BusDomain, sd), "query")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unittest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsers(ctx, 1, role.Admin, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	audits, err := auditbus.TestSeedAudits(ctx, 2, usrs[0].ID, domain.User, "create", busDomain.Audit)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := unittest.User{
		User:   usrs[0],
		Audits: audits,
	}

	// -------------------------------------------------------------------------

	sd := unittest.SeedData{
		Admins: []unittest.User{tu1},
	}

	return sd, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	sort.Slice(sd.Admins[0].Audits, func(i, j int) bool {
		return sd.Admins[0].Audits[i].ObjName.String() <= sd.Admins[0].Audits[j].ObjName.String()
	})

	table := []unittest.Table{
		{
			Name:    "all",
			ExpResp: sd.Admins[0].Audits,
			ExcFunc: func(ctx context.Context) any {
				filter := auditbus.QueryFilter{
					Action: dbtest.StringPointer("create"),
				}

				orderBy := order.NewBy(auditbus.OrderByObjName, order.ASC)

				resp, err := busDomain.Audit.Query(ctx, filter, orderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]auditbus.Audit)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]auditbus.Audit)

				for i := range gotResp {
					if gotResp[i].Timestamp.Format(time.RFC3339) == expResp[i].Timestamp.Format(time.RFC3339) {
						expResp[i].Timestamp = gotResp[i].Timestamp
					}

					gotResp[i].Data = bytes.ReplaceAll(gotResp[i].Data, []byte{' '}, []byte{})
					expResp[i].Data = bytes.ReplaceAll(expResp[i].Data, []byte{' '}, []byte{})
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}
