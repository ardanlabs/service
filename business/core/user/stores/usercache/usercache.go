// Package usercache contains user related CRUD functionality with caching.
package usercache

import (
	"context"
	"net/mail"
	"sync"

	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/data/transaction"
	"github.com/ardanlabs/service/business/web/v1/order"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/google/uuid"
)

// Store manages the set of APIs for user data and caching.
type Store struct {
	log    *logger.Logger
	storer user.Storer
	cache  map[string]user.User
	mu     sync.RWMutex
}

// NewStore constructs the api for data and caching access.
func NewStore(log *logger.Logger, storer user.Storer) *Store {
	return &Store{
		log:    log,
		storer: storer,
		cache:  map[string]user.User{},
	}
}

// ExecuteUnderTransaction constructs a new Store value replacing the sqlx DB
// value with a sqlx DB value that is currently inside a transaction.
func (s *Store) ExecuteUnderTransaction(tx transaction.Transaction) (user.Storer, error) {
	return s.storer.ExecuteUnderTransaction(tx)
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	if err := s.storer.Create(ctx, usr); err != nil {
		return err
	}

	s.writeCache(usr)

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	if err := s.storer.Update(ctx, usr); err != nil {
		return err
	}

	s.writeCache(usr)

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	if err := s.storer.Delete(ctx, usr); err != nil {
		return err
	}

	s.deleteCache(usr)

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]user.User, error) {
	return s.storer.Query(ctx, filter, orderBy, pageNumber, rowsPerPage)
}

// Count returns the total number of cards in the DB.
func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	return s.storer.Count(ctx, filter)
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	cachedUsr, ok := s.readCache(userID.String())
	if ok {
		return cachedUsr, nil
	}

	usr, err := s.storer.QueryByID(ctx, userID)
	if err != nil {
		return user.User{}, err
	}

	s.writeCache(usr)

	return usr, nil
}

// QueryByIDs gets the specified users from the database.
func (s *Store) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	usr, err := s.storer.QueryByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	cachedUsr, ok := s.readCache(email.Address)
	if ok {
		return cachedUsr, nil
	}

	usr, err := s.storer.QueryByEmail(ctx, email)
	if err != nil {
		return user.User{}, err
	}

	s.writeCache(usr)

	return usr, nil
}

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (user.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	usr, exists := s.cache[key]
	if !exists {
		return user.User{}, false
	}

	return usr, true
}

// writeCache performs a safe write to the cache for the specified user.
func (s *Store) writeCache(usr user.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[usr.ID.String()] = usr
	s.cache[usr.Email.Address] = usr
}

// deleteCache performs a safe removal from the cache for the specified user.
func (s *Store) deleteCache(usr user.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, usr.ID.String())
	delete(s.cache, usr.Email.Address)
}
