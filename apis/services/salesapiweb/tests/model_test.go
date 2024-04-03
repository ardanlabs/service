package tests

import (
	"time"

	"github.com/ardanlabs/service/app/core/crud/homeapp"
	"github.com/ardanlabs/service/app/core/crud/productapp"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/app/core/views/vproductapp"
	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
)

func toErrorPtr(err errs.Error) *errs.Error {
	return &err
}

func toAppUser(usr user.User) userapp.User {
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

func toAppUsers(users []user.User) []userapp.User {
	items := make([]userapp.User, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(usr user.User) *userapp.User {
	appUsr := toAppUser(usr)
	return &appUsr
}

func toAppProduct(prd product.Product) productapp.Product {
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

func toAppProductPtr(prd product.Product) *productapp.Product {
	appPrd := toAppProduct(prd)
	return &appPrd
}

func toAppProducts(prds []product.Product) []productapp.Product {
	items := make([]productapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}

func toAppHome(hme home.Home) homeapp.Home {
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

func toAppHomes(homes []home.Home) []homeapp.Home {
	items := make([]homeapp.Home, len(homes))
	for i, hme := range homes {
		items[i] = toAppHome(hme)
	}

	return items
}

func toAppHomePtr(hme home.Home) *homeapp.Home {
	appHme := toAppHome(hme)
	return &appHme
}

func toAppVProduct(usr user.User, prd product.Product) vproductapp.Product {
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

func toAppVProducts(usr user.User, prds []product.Product) []vproductapp.Product {
	items := make([]vproductapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppVProduct(usr, prd)
	}

	return items
}
