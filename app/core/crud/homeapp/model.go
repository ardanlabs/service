package homeapp

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/api/errs"
	"github.com/ardanlabs/service/business/api/mid"
	"github.com/ardanlabs/service/business/core/crud/homebus"
	"github.com/ardanlabs/service/foundation/validate"
)

// QueryParams represents the set of possible query strings.
type QueryParams struct {
	Page             int    `query:"page"`
	Rows             int    `query:"rows"`
	OrderBy          string `query:"orderBy"`
	ID               string `query:"home_id"`
	UserID           string `query:"user_id"`
	Type             string `query:"type"`
	StartCreatedDate string `query:"start_created_date"`
	EndCreatedDate   string `query:"end_created_date"`
}

// Address represents information about an individual address.
type Address struct {
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	ZipCode  string `json:"zipCode"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
}

// Home represents information about an individual home.
type Home struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Type        string  `json:"type"`
	Address     Address `json:"address"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

func toAppHome(hme homebus.Home) Home {
	return Home{
		ID:     hme.ID.String(),
		UserID: hme.UserID.String(),
		Type:   hme.Type.Name(),
		Address: Address{
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

func toAppHomes(homes []homebus.Home) []Home {
	items := make([]Home, len(homes))
	for i, hme := range homes {
		items[i] = toAppHome(hme)
	}

	return items
}

// NewAddress defines the data needed to add a new address.
type NewAddress struct {
	Address1 string `json:"address1" validate:"required,min=1,max=70"`
	Address2 string `json:"address2" validate:"omitempty,max=70"`
	ZipCode  string `json:"zipCode" validate:"required,numeric"`
	City     string `json:"city" validate:"required"`
	State    string `json:"state" validate:"required,min=1,max=48"`
	Country  string `json:"country" validate:"required,iso3166_1_alpha2"`
}

// NewHome defines the data needed to add a new home.
type NewHome struct {
	Type    string     `json:"type" validate:"required"`
	Address NewAddress `json:"address"`
}

func toBusNewHome(ctx context.Context, app NewHome) (homebus.NewHome, error) {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return homebus.NewHome{}, fmt.Errorf("getuserid: %w", err)
	}

	typ, err := homebus.ParseType(app.Type)
	if err != nil {
		return homebus.NewHome{}, fmt.Errorf("parse: %w", err)
	}

	hme := homebus.NewHome{
		UserID: userID,
		Type:   typ,
		Address: homebus.Address{
			Address1: app.Address.Address1,
			Address2: app.Address.Address2,
			ZipCode:  app.Address.ZipCode,
			City:     app.Address.City,
			State:    app.Address.State,
			Country:  app.Address.Country,
		},
	}

	return hme, nil
}

// Validate checks if the data in the model is considered clean.
func (app NewHome) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}

// UpdateAddress defines the data needed to update an address.
type UpdateAddress struct {
	Address1 *string `json:"address1" validate:"omitempty,min=1,max=70"`
	Address2 *string `json:"address2" validate:"omitempty,max=70"`
	ZipCode  *string `json:"zipCode" validate:"omitempty,numeric"`
	City     *string `json:"city"`
	State    *string `json:"state" validate:"omitempty,min=1,max=48"`
	Country  *string `json:"country" validate:"omitempty,iso3166_1_alpha2"`
}

// Validate checks the data in the model is considered clean.
func (app UpdateAddress) Validate() error {
	if err := validate.Check(app); err != nil {
		return err
	}

	return nil
}

// UpdateHome defines the data needed to update a home.
type UpdateHome struct {
	Type    *string        `json:"type"`
	Address *UpdateAddress `json:"address"`
}

func toBusUpdateHome(app UpdateHome) (homebus.UpdateHome, error) {
	var typ homebus.Type
	if app.Type != nil {
		var err error
		typ, err = homebus.ParseType(*app.Type)
		if err != nil {
			return homebus.UpdateHome{}, fmt.Errorf("parse: %w", err)
		}
	}

	core := homebus.UpdateHome{
		Type: &typ,
	}

	if app.Address != nil {
		core.Address = &homebus.UpdateAddress{
			Address1: app.Address.Address1,
			Address2: app.Address.Address2,
			ZipCode:  app.Address.ZipCode,
			City:     app.Address.City,
			State:    app.Address.State,
			Country:  app.Address.Country,
		}
	}

	return core, nil
}

// Validate checks the data in the model is considered clean.
func (app UpdateHome) Validate() error {
	if err := validate.Check(app); err != nil {
		return errs.Newf(errs.FailedPrecondition, "validate: %s", err)
	}

	return nil
}
