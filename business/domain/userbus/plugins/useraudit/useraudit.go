// Package useraudit provides a plugin for userbus that adds
// auditing functionality.
package useraudit

import (
	"context"
	"net/mail"

	"github.com/ardanlabs/service/business/domain/auditbus"
	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/business/types/domain"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// Plugin provides a wrapper for audit functionality around the userbus.
type Plugin struct {
	bus      userbus.Business
	auditBus *auditbus.Business
}

// NewPlugin constructs a new plugin that wraps the userbus with audit.
func NewPlugin(log *logger.Logger, auditBus *auditbus.Business) userbus.Plugin {
	return func(bus userbus.Business) userbus.Business {
		return &Plugin{
			bus:      bus,
			auditBus: auditBus,
		}
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (p *Plugin) NewWithTx(tx sqldb.CommitRollbacker) (userbus.Business, error) {
	return p.bus.NewWithTx(tx)
}

// Create adds a new user to the system.
func (p *Plugin) Create(ctx context.Context, actorID uuid.UUID, nu userbus.NewUser) (userbus.User, error) {
	usr, err := p.bus.Create(ctx, actorID, nu)
	if err != nil {
		return userbus.User{}, err
	}

	na := auditbus.NewAudit{
		ObjID:     usr.ID,
		ObjDomain: domain.User,
		ObjName:   usr.Name,
		ActorID:   actorID,
		Action:    "created",
		Data:      nu,
		Message:   "user created",
	}

	if _, err := p.auditBus.Create(ctx, na); err != nil {
		return userbus.User{}, err
	}

	return usr, nil
}

// Update modifies information about a user.
func (p *Plugin) Update(ctx context.Context, actorID uuid.UUID, usr userbus.User, uu userbus.UpdateUser) (userbus.User, error) {
	usr, err := p.bus.Update(ctx, actorID, usr, uu)
	if err != nil {
		return userbus.User{}, err
	}

	na := auditbus.NewAudit{
		ObjID:     usr.ID,
		ObjDomain: domain.User,
		ObjName:   usr.Name,
		ActorID:   actorID,
		Action:    "updated",
		Data:      uu,
		Message:   "user updated",
	}

	if _, err := p.auditBus.Create(ctx, na); err != nil {
		return userbus.User{}, err
	}

	return usr, nil
}

// Delete removes the specified user.
func (p *Plugin) Delete(ctx context.Context, actorID uuid.UUID, usr userbus.User) error {
	if err := p.bus.Delete(ctx, actorID, usr); err != nil {
		return err
	}

	na := auditbus.NewAudit{
		ObjID:     usr.ID,
		ObjDomain: domain.User,
		ObjName:   usr.Name,
		ActorID:   actorID,
		Action:    "deleted",
		Data:      nil,
		Message:   "user deleted",
	}

	if _, err := p.auditBus.Create(ctx, na); err != nil {
		return err
	}

	return nil
}

// Query retrieves a list of existing users.
func (p *Plugin) Query(ctx context.Context, filter userbus.QueryFilter, orderBy order.By, page page.Page) ([]userbus.User, error) {
	return p.bus.Query(ctx, filter, orderBy, page)
}

// Count returns the total number of users.
func (p *Plugin) Count(ctx context.Context, filter userbus.QueryFilter) (int, error) {
	return p.bus.Count(ctx, filter)
}

// QueryByID finds the user by the specified ID.
func (p *Plugin) QueryByID(ctx context.Context, userID uuid.UUID) (userbus.User, error) {
	return p.bus.QueryByID(ctx, userID)
}

// QueryByEmail finds the user by a specified user email.
func (p *Plugin) QueryByEmail(ctx context.Context, email mail.Address) (userbus.User, error) {
	return p.bus.QueryByEmail(ctx, email)
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (p *Plugin) Authenticate(ctx context.Context, email mail.Address, password string) (userbus.User, error) {
	return p.bus.Authenticate(ctx, email, password)
}
