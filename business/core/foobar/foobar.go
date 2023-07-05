package foobar

import (
	"context"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/sys/logger"
	"github.com/ardanlabs/service/foundation/core"
)

// Core manages the set of APIs for product access.
type Core struct {
	log     *logger.Logger
	usrCore *user.Core
	prdCore *product.Core
}

// NewCore constructs a core for product api access.
func NewCore(log *logger.Logger, usrCore *user.Core, prdCore *product.Core) *Core {
	core := Core{
		log:     log,
		usrCore: usrCore,
		prdCore: prdCore,
	}

	return &core
}

func (c *Core) Begin() (core.Transactor, error) {
	return c.usrCore.Begin()
}

func (c *Core) InTran(tr core.Transactor) (*Core, error) {
	usrCore, err := c.usrCore.InTran(tr)
	if err != nil {
		return nil, err
	}
	prdCore, err := c.prdCore.InTran(tr)
	if err != nil {
		return nil, err
	}
	return &Core{
		log:     c.log,
		usrCore: usrCore,
		prdCore: prdCore,
	}, nil
}

func (c *Core) Create(ctx context.Context, np product.NewProduct, nu user.NewUser) (product.Product, error) {
	var prd product.Product
	tran := func(c *Core) error {
		usr, err := c.usrCore.Create(ctx, nu)
		if err != nil {
			return err
		}

		np.UserID = usr.ID
		prd, err = c.prdCore.Create(ctx, np)
		if err != nil {
			return err
		}
		return nil
	}

	if err := core.WithinTranCore[*Core](ctx, c.log, c, tran); err != nil {
		return prd, err
	}
	return prd, nil
}
