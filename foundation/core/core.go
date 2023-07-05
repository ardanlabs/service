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

type Beginner[T any] interface {
	Begin() (Transactor, error)
	InTran(Transactor) (T, error)
}

type StoreBeginner[T any] interface {
	Begin() (Transactor, error)
	InTran(Transactor) (T, error)
	IsInTran() bool
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
	return WithinTranFn(ctx, log, tr, tran)
}

func WithinTranStore[T any](ctx context.Context, log *logger.Logger, b StoreBeginner[T], s T, fn func(s T) error) error {
	if b.IsInTran() {
		return fn(s)
	}
	tr, err := b.Begin()
	if err != nil {
		return err
	}

	tran := func(tr Transactor) error {
		trS, err := b.InTran(tr)
		if err != nil {
			return err
		}
		if err := fn(trS); err != nil {
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
