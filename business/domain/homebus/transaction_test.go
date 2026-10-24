package homebus

import (
	"bytes"
	"context"
	"testing"

	"github.com/ardanlabs/service/business/domain/userbus"
	"github.com/ardanlabs/service/business/sdk/delegate"
	"github.com/ardanlabs/service/business/sdk/sqldb"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

func TestNewWithTxDoesNotRegisterDelegate(t *testing.T) {
	var callbackCalls int
	log := logger.NewWithEvents(&bytes.Buffer{}, logger.LevelInfo, "TEST", nil, logger.Events{
		Info: func(ctx context.Context, r logger.Record) {
			if r.Message == "action-userdeleted" {
				callbackCalls++
			}
		},
	})
	d := delegate.New(log)

	extension := func(bus ExtBusiness) ExtBusiness {
		return &testExtension{ExtBusiness: bus}
	}

	bus := NewBusiness(log, testUserBusiness{}, d, testStorer{}, extension)
	for range 2 {
		var err error
		bus, err = bus.NewWithTx(testTransaction{})
		if err != nil {
			t.Fatalf("NewWithTx: %s", err)
		}

		count, err := bus.Count(t.Context(), QueryFilter{})
		if err != nil {
			t.Fatalf("Count: %s", err)
		}
		if count != 11 {
			t.Fatalf("count with extension: got %d, want 11", count)
		}
	}

	if err := d.Call(t.Context(), userbus.ActionDeletedData(uuid.New())); err != nil {
		t.Fatalf("Call: %s", err)
	}

	if callbackCalls != 1 {
		t.Fatalf("callback calls: got %d, want 1", callbackCalls)
	}
}

type testStorer struct {
	Storer
}

func (testStorer) NewWithTx(sqldb.CommitRollbacker) (Storer, error) {
	return testStorer{}, nil
}

func (testStorer) Count(context.Context, QueryFilter) (int, error) {
	return 10, nil
}

type testUserBusiness struct {
	userbus.ExtBusiness
}

func (testUserBusiness) NewWithTx(sqldb.CommitRollbacker) (userbus.ExtBusiness, error) {
	return testUserBusiness{}, nil
}

type testExtension struct {
	ExtBusiness
}

func (ext *testExtension) NewWithTx(tx sqldb.CommitRollbacker) (ExtBusiness, error) {
	bus, err := ext.ExtBusiness.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	next := testExtension{ExtBusiness: bus}
	return &next, nil
}

func (ext *testExtension) Count(ctx context.Context, filter QueryFilter) (int, error) {
	count, err := ext.ExtBusiness.Count(ctx, filter)
	return count + 1, err
}

type testTransaction struct{}

func (testTransaction) Commit() error {
	return nil
}

func (testTransaction) Rollback() error {
	return nil
}
