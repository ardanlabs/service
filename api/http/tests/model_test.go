package tests

import (
	"time"

	"github.com/ardanlabs/service/app/api/errs"
	"github.com/ardanlabs/service/app/domain/homeapp"
	"github.com/ardanlabs/service/app/domain/productapp"
	"github.com/ardanlabs/service/app/domain/userapp"
	"github.com/ardanlabs/service/app/domain/vproductapp"
	"github.com/ardanlabs/service/business/domain/homebus"
	"github.com/ardanlabs/service/business/domain/productbus"
	"github.com/ardanlabs/service/business/domain/userbus"
)

func toErrorPtr(err errs.Error) *errs.Error {
	return &err
}

func toAppUser(usr userbus.User) userapp.User {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return userapp.User{
		ID:           usr.ID.String(),
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        roles,
		PasswordHash: nil, // This field is not marshalled.
		Department:   usr.Department,
		Enabled:      usr.Enabled,
		DateCreated:  usr.DateCreated.Format(time.RFC3339),
		DateUpdated:  usr.DateUpdated.Format(time.RFC3339),
	}
}

func toAppUsers(users []userbus.User) []userapp.User {
	items := make([]userapp.User, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(usr userbus.User) *userapp.User {
	appUsr := toAppUser(usr)
	return &appUsr
}

func toAppProduct(prd productbus.Product) productapp.Product {
	return productapp.Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProductPtr(prd productbus.Product) *productapp.Product {
	appPrd := toAppProduct(prd)
	return &appPrd
}

func toAppProducts(prds []productbus.Product) []productapp.Product {
	items := make([]productapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}

func toAppHome(hme homebus.Home) homeapp.Home {
	return homeapp.Home{
		ID:     hme.ID.String(),
		UserID: hme.UserID.String(),
		Type:   hme.Type.Name(),
		Address: homeapp.Address{
			Address1: hme.Address.Address1,
			Address2: hme.Address.Address2,
			ZipCode:  hme.Address.ZipCode,
			City:     hme.Address.City,
			State:    hme.Address.State,
			Country:  hme.Address.Country,
		},
		DateCreated: hme.DateCreated.Format(time.RFC3339),
		DateUpdated: hme.DateUpdated.Format(time.RFC3339),
	}
}

func toAppHomes(homes []homebus.Home) []homeapp.Home {
	items := make([]homeapp.Home, len(homes))
	for i, hme := range homes {
		items[i] = toAppHome(hme)
	}

	return items
}

func toAppHomePtr(hme homebus.Home) *homeapp.Home {
	appHme := toAppHome(hme)
	return &appHme
}

func toAppVProduct(usr userbus.User, prd productbus.Product) vproductapp.Product {
	return vproductapp.Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
		UserName:    usr.Name,
	}
}

func toAppVProducts(usr userbus.User, prds []productbus.Product) []vproductapp.Product {
	items := make([]vproductapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppVProduct(usr, prd)
	}

	return items
}
