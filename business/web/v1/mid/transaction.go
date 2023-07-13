package mid

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/sys/core"
	"github.com/ardanlabs/service/foundation/web"
)

// func InTran(db *sqlx.DB) web.Middleware {
func InTran(f core.BeginnerFactory) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			tx, err := f.Begin()
			if err != nil {
				return err
			}
			defer func() {
				if err := tx.Rollback(); err != nil {
					if errors.Is(err, sql.ErrTxDone) {
						return
					}
				}
			}()
			ctx = core.SetTransactor(ctx, tx)
			err = handler(ctx, w, r)
			if err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("commit tran: %w", err)
			}
			return nil
		}

		return h
	}

	return m
}
