package tests

import (
	"time"

	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/homegrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/productgrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/usergrp"
	"github.com/ardanlabs/service/app/services/sales-api/v1/handlers/vproductgrp"
	"github.com/ardanlabs/service/business/core/crud/home"
	"github.com/ardanlabs/service/business/core/crud/product"
	"github.com/ardanlabs/service/business/core/crud/user"
)

type tableData struct {
	name       string
	url        string
	token      string
	method     string
	statusCode int
	model      any
	resp       any
	expResp    any
	cmpFunc    func(x interface{}, y interface{}) string
}

type testUser struct {
	user.User
	token    string
	products []product.Product
	homes    []home.Home
}

type seedData struct {
	users  []testUser
	admins []testUser
}

func toAppUser(usr user.User) usergrp.AppUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return usergrp.AppUser{
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

func toAppUsers(users []user.User) []usergrp.AppUser {
	items := make([]usergrp.AppUser, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(usr user.User) *usergrp.AppUser {
	appUsr := toAppUser(usr)
	return &appUsr
}

func toAppProduct(prd product.Product) productgrp.AppProduct {
	return productgrp.AppProduct{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name,
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

func toAppProductPtr(prd product.Product) *productgrp.AppProduct {
	appPrd := toAppProduct(prd)
	return &appPrd
}

func toAppProducts(prds []product.Product) []productgrp.AppProduct {
	items := make([]productgrp.AppProduct, len(prds))
	for i, prd := range prds {
		items[i] = toAppProduct(prd)
	}

	return items
}

func toAppHome(hme home.Home) homegrp.AppHome {
	return homegrp.AppHome{
		ID:     hme.ID.String(),
		UserID: hme.UserID.String(),
		Type:   hme.Type.Name(),
		Address: homegrp.AppAddress{
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

func toAppHomes(homes []home.Home) []homegrp.AppHome {
	items := make([]homegrp.AppHome, len(homes))
	for i, hme := range homes {
		items[i] = toAppHome(hme)
	}

	return items
}

func toAppHomePtr(hme home.Home) *homegrp.AppHome {
	appHme := toAppHome(hme)
	return &appHme
}

func toAppVProduct(usr user.User, prd product.Product) vproductgrp.AppProduct {
	return vproductgrp.AppProduct{
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

func toAppVProducts(usr user.User, prds []product.Product) []vproductgrp.AppProduct {
	items := make([]vproductgrp.AppProduct, len(prds))
	for i, prd := range prds {
		items[i] = toAppVProduct(usr, prd)
	}

	return items
}
