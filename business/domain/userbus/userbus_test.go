package userbus_test

import (
	"context"
	"fmt"
	"net/mail"
	"sort"
	"testing"
	"time"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/dbtest"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/unittest"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/password"
	"github.com/ardanlabs/service/business/types/role"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Test_User(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_User")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// -------------------------------------------------------------------------

	unittest.Run(t, query(db.BusDomain, sd), "query")
	unittest.Run(t, create(db.BusDomain), "create")
	unittest.Run(t, update(db.BusDomain, sd), "update")
	unittest.Run(t, delete(db.BusDomain, sd), "delete")
}

// =============================================================================

func insertSeedData(busDomain dbtest.BusDomain) (unittest.SeedData, error) {
	ctx := context.Background()

	usrs, err := userbus.TestSeedUsers(ctx, 2, role.Admin, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu1 := unittest.User{
		User: usrs[0],
	}

	tu2 := unittest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	usrs, err = userbus.TestSeedUsers(ctx, 2, role.User, busDomain.User)
	if err != nil {
		return unittest.SeedData{}, fmt.Errorf("seeding users : %w", err)
	}

	tu3 := unittest.User{
		User: usrs[0],
	}

	tu4 := unittest.User{
		User: usrs[1],
	}

	// -------------------------------------------------------------------------

	sd := unittest.SeedData{
		Users:  []unittest.User{tu3, tu4},
		Admins: []unittest.User{tu1, tu2},
	}

	return sd, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
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

	table := []unittest.Table{
		{
			Name:    "all",
			ExpResp: usrs,
			ExcFunc: func(ctx context.Context) any {
				filter := userbus.QueryFilter{
					Name: dbtest.NamePointer("Name"),
				}

				resp, err := busDomain.User.Query(ctx, filter, userbus.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.([]userbus.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.([]userbus.User)

				for i := range gotResp {
					if gotResp[i].DateCreated.Format(time.RFC3339) == expResp[i].DateCreated.Format(time.RFC3339) {
						expResp[i].DateCreated = gotResp[i].DateCreated
					}

					if gotResp[i].DateUpdated.Format(time.RFC3339) == expResp[i].DateUpdated.Format(time.RFC3339) {
						expResp[i].DateUpdated = gotResp[i].DateUpdated
					}
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Users[0].User,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.User.QueryByID(ctx, sd.Users[0].ID)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				expResp := exp.(userbus.User)

				if gotResp.DateCreated.Format(time.RFC3339) == expResp.DateCreated.Format(time.RFC3339) {
					expResp.DateCreated = gotResp.DateCreated
				}

				if gotResp.DateUpdated.Format(time.RFC3339) == expResp.DateUpdated.Format(time.RFC3339) {
					expResp.DateUpdated = gotResp.DateUpdated
				}

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func create(busDomain dbtest.BusDomain) []unittest.Table {
	email, _ := mail.ParseAddress("bill@ardanlabs.com")

	table := []unittest.Table{
		{
			Name: "basic",
			ExpResp: userbus.User{
				Name:       name.MustParse("Bill Kennedy"),
				Email:      *email,
				Roles:      []role.Role{role.Admin},
				Department: name.MustParseNull("ITO"),
				Enabled:    true,
			},
			ExcFunc: func(ctx context.Context) any {
				nu := userbus.NewUser{
					Name:       name.MustParse("Bill Kennedy"),
					Email:      *email,
					Roles:      []role.Role{role.Admin},
					Department: name.MustParseNull("ITO"),
					Password:   password.MustParse("123"),
				}

				resp, err := busDomain.User.Create(ctx, uuid.UUID{}, nu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("123")); err != nil {
					return err.Error()
				}

				expResp := exp.(userbus.User)

				expResp.ID = gotResp.ID
				expResp.PasswordHash = gotResp.PasswordHash
				expResp.DateCreated = gotResp.DateCreated
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func update(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	email, _ := mail.ParseAddress("jack@ardanlabs.com")

	table := []unittest.Table{
		{
			Name: "basic",
			ExpResp: userbus.User{
				ID:          sd.Users[0].ID,
				Name:        name.MustParse("Jack Kennedy"),
				Email:       *email,
				Roles:       []role.Role{role.Admin},
				Department:  name.MustParseNull("ITO"),
				Enabled:     true,
				DateCreated: sd.Users[0].DateCreated,
			},
			ExcFunc: func(ctx context.Context) any {
				uu := userbus.UpdateUser{
					Name:       dbtest.NamePointer("Jack Kennedy"),
					Email:      email,
					Roles:      []role.Role{role.Admin},
					Department: dbtest.NameNullPointer("ITO"),
					Password:   dbtest.PasswordPointer("1234"),
				}

				resp, err := busDomain.User.Update(ctx, uuid.UUID{}, sd.Users[0].User, uu)
				if err != nil {
					return err
				}

				return resp
			},
			CmpFunc: func(got any, exp any) string {
				gotResp, exists := got.(userbus.User)
				if !exists {
					return "error occurred"
				}

				if err := bcrypt.CompareHashAndPassword(gotResp.PasswordHash, []byte("1234")); err != nil {
					return err.Error()
				}

				expResp := exp.(userbus.User)

				expResp.PasswordHash = gotResp.PasswordHash
				expResp.DateUpdated = gotResp.DateUpdated

				return cmp.Diff(gotResp, expResp)
			},
		},
	}

	return table
}

func delete(busDomain dbtest.BusDomain, sd unittest.SeedData) []unittest.Table {
	table := []unittest.Table{
		{
			Name:    "user",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, uuid.UUID{}, sd.Users[1].User); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "admin",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.User.Delete(ctx, uuid.UUID{}, sd.Admins[1].User); err != nil {
					return err
				}

				return nil
			},
			CmpFunc: func(got any, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}

	return table
}
