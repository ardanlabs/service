// Package usercache contains user related CRUD functionality with caching.
package usercache

import (
	"context"
	"errors"

	"github.com/ardanlabs/service/business/core/user"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user data and caching.
type Store struct {
	log    *zap.SugaredLogger
	storer user.Storer
	cache  map[string]*user.User
}

// NewStore constructs the api for data and caching access.
func NewStore(log *zap.SugaredLogger, storer user.Storer) Store {
	return Store{
		log:    log,
		storer: storer,
		cache:  map[string]*user.User{},
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s Store) WithinTran(ctx context.Context, fn func(s user.Storer) error) error {
	return s.storer.WithinTran(ctx, fn)
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, usr user.User) error {
	if err := s.storer.Create(ctx, usr); err != nil {
		return err
	}

	s.cache[usr.ID] = &usr
	s.cache[usr.Email] = &usr

	return nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, usr user.User) error {
	if err := s.storer.Update(ctx, usr); err != nil {
		return err
	}

	s.cache[usr.ID] = &usr
	s.cache[usr.Email] = &usr

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, userID string) error {
	usr, err := s.storer.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return nil
		}
		return err
	}

	if err := s.storer.Delete(ctx, usr.ID); err != nil {
		return err
	}

	delete(s.cache, usr.ID)
	delete(s.cache, usr.Email)

	return nil
}

// Query retrieves a list of existing users from the database.
func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]user.User, error) {
	return s.storer.Query(ctx, pageNumber, rowsPerPage)
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, userID string) (user.User, error) {
	cachedUsr, ok := s.cache[userID]
	if ok {
		return *cachedUsr, nil
	}

	usr, err := s.storer.QueryByID(ctx, userID)
	if err != nil {
		return user.User{}, err
	}

	s.cache[usr.ID] = &usr
	s.cache[usr.Email] = &usr

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s Store) QueryByEmail(ctx context.Context, email string) (user.User, error) {
	cachedUsr, ok := s.cache[email]
	if ok {
		return *cachedUsr, nil
	}

	usr, err := s.storer.QueryByEmail(ctx, email)
	if err != nil {
		return user.User{}, err
	}

	s.cache[usr.ID] = &usr
	s.cache[usr.Email] = &usr

	return usr, nil
}
