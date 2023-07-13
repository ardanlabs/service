package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ardanlabs/service/business/sys/logger"
)

type Transactor interface {
	Commit() error
	Rollback() error
}

type InTraner[T any] interface {
	InTran(Transactor) (T, error)
}

type BeginnerFactory interface {
	Begin() (Transactor, error)
}

func WithinTranCore[T any](ctx context.Context, log *logger.Logger, tr Transactor, b InTraner[T], fn func(c T) error) error {
	tran := func(tr Transactor) error {
		trCore, err := b.InTran(tr)
		if err != nil {
			return err
		}
		if err := fn(trCore); err != nil {
			return err
		}
		return nil
	}
	return WithinTranFn(ctx, log, tr, tran)
}

func WithinTranFn(ctx context.Context, log *logger.Logger, tr Transactor, fn func(tr Transactor) error) error {
	defer func() {
		if err := tr.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Error(ctx, "unable to rollback tran", "msg", err)
		}
		log.Info(ctx, "rollback tran")
	}()

	if err := fn(tr); err != nil {
		return fmt.Errorf("exec tran: %w", err)
	}

	if err := tr.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}
	return nil
}

type ctxKey int

const trKey ctxKey = 2

func SetTransactor(ctx context.Context, tr Transactor) context.Context {
	return context.WithValue(ctx, trKey, tr)
}

func GetTransactor(ctx context.Context) (Transactor, bool) {
	v, ok := ctx.Value(trKey).(Transactor)
	return v, ok
}
