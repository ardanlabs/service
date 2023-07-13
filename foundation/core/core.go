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

type NestedTransactor interface {
	Transactor
	IsNested() bool
}

type Beginner[T any] interface {
	Begin() (NestedTransactor, error)
	InTran(Transactor) (T, error)
}

func WithinTranCore[T any](ctx context.Context, log *logger.Logger, b Beginner[T], fn func(c T) error) error {
	tr, err := b.Begin()
	if err != nil {
		return err
	}

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
	if !tr.IsNested() {
		return WithinTranFn(ctx, log, tr, tran)
	}
	return tran(tr)
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

type NestedTransaction struct {
	Tr     Transactor
	Nested bool
}

func (nr *NestedTransaction) Commit() error {
	return nr.Tr.Commit()
}

func (nr *NestedTransaction) Rollback() error {
	return nr.Tr.Rollback()
}

func (nr *NestedTransaction) IsNested() bool {
	return nr.Nested
}
