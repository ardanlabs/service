package sturdyc

import (
	"math/rand/v2"
	"sync"
	"time"
)

// entry represents a single cache entry.
type entry[T any] struct {
	key                 string
	value               T
	expiresAt           time.Time
	refreshAt           time.Time
	numOfRefreshRetries int
	isMissingRecord     bool
}

// shard is a thread-safe data structure that holds a subset of the cache entries.
type shard[T any] struct {
	sync.RWMutex
	*Config
	capacity           int
	ttl                time.Duration
	entries            map[string]*entry[T]
	evictionPercentage int
}

// newShard creates a new shard and returns a pointer to it.
func newShard[T any](capacity int, ttl time.Duration, evictionPercentage int, cfg *Config) *shard[T] {
	return &shard[T]{
		Config:             cfg,
		capacity:           capacity,
		ttl:                ttl,
		entries:            make(map[string]*entry[T]),
		evictionPercentage: evictionPercentage,
	}
}

// size returns the number of entries in the shard.
func (s *shard[T]) size() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.entries)
}

// evictExpired evicts all the expired entries in the shard.
func (s *shard[T]) evictExpired() {
	s.Lock()
	defer s.Unlock()

	var entriesEvicted int
	for _, e := range s.entries {
		if s.clock.Now().After(e.expiresAt) {
			delete(s.entries, e.key)
			entriesEvicted++
		}
	}
	s.reportEntriesEvicted(entriesEvicted)
}

// forceEvict evicts a certain percentage of the entries in the shard
// based on the expiration time. Should be called with a lock.
func (s *shard[T]) forceEvict() {
	s.reportForcedEviction()
	expirationTimes := make([]time.Time, 0, len(s.entries))
	for _, e := range s.entries {
		expirationTimes = append(expirationTimes, e.expiresAt)
	}

	cutoff := FindCutoff(expirationTimes, float64(s.evictionPercentage)/100)
	entriesEvicted := 0
	for key, e := range s.entries {
		if e.expiresAt.Before(cutoff) {
			delete(s.entries, key)
			entriesEvicted++
		}
	}
	s.reportEntriesEvicted(entriesEvicted)
}

// get retrieves attempts to retrieve a value from the shard.
//
// Parameters:
//
//	key: The key for which the value is to be retrieved.
//
// Returns:
//
//	val: The value associated with the key, if it exists.
//	exists: A boolean indicating if the value exists in the shard.
//	markedAsMissing: A boolean indicating if the key has been marked as a missing record.
//	refresh: A boolean indicating if the value should be refreshed in the background.
func (s *shard[T]) get(key string) (val T, exists, markedAsMissing, refresh bool) {
	s.RLock()
	item, ok := s.entries[key]
	if !ok {
		s.RUnlock()
		return val, false, false, false
	}

	if s.clock.Now().After(item.expiresAt) {
		s.RUnlock()
		return val, false, false, false
	}

	shouldRefresh := s.refreshInBackground && s.clock.Now().After(item.refreshAt)
	if shouldRefresh {
		// Release the read lock, and switch to a write lock.
		s.RUnlock()
		s.Lock()

		// However, during the time it takes to switch locks, another goroutine
		// might have acquired it and moved the refreshAt. Therefore, we'll have to
		// check if this operation should still be performed.
		if !s.clock.Now().After(item.refreshAt) {
			s.Unlock()
			return item.value, true, item.isMissingRecord, false
		}

		// Update the "refreshAt" so no other goroutines attempts to refresh the same entry.
		nextRefresh := s.retryBaseDelay * (1 << item.numOfRefreshRetries)
		item.refreshAt = s.clock.Now().Add(nextRefresh)
		item.numOfRefreshRetries++

		s.Unlock()
		return item.value, true, item.isMissingRecord, shouldRefresh
	}

	s.RUnlock()
	return item.value, true, item.isMissingRecord, false
}

// set writes a key-value pair to the shard and returns a
// boolean indicating whether an eviction was performed.
func (s *shard[T]) set(key string, value T, isMissingRecord bool) bool {
	s.Lock()
	defer s.Unlock()

	// Check we need to perform an eviction first.
	evict := len(s.entries) >= s.capacity

	// If the cache is configured to not evict any entries,
	// and we're att full capacity, we'll return early.
	if s.evictionPercentage < 1 && evict {
		return false
	}

	if evict {
		s.forceEvict()
	}

	now := s.clock.Now()
	newEntry := &entry[T]{
		key:             key,
		value:           value,
		expiresAt:       now.Add(s.ttl),
		isMissingRecord: isMissingRecord,
	}

	if s.refreshInBackground {
		// If there is a difference between the min- and maxRefreshTime we'll use that to
		// set a random padding so that the refreshes get spread out evenly over time.
		var padding time.Duration
		if s.minRefreshTime != s.maxRefreshTime {
			padding = time.Duration(rand.Int64N(int64(s.maxRefreshTime - s.minRefreshTime)))
		}
		newEntry.refreshAt = now.Add(s.minRefreshTime + padding)
		newEntry.numOfRefreshRetries = 0
	}

	s.entries[key] = newEntry
	return evict
}

// delete removes a key from the shard.
func (s *shard[T]) delete(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.entries, key)
}

// keys returns all non-expired keys in the shard.
func (s *shard[T]) keys() []string {
	s.RLock()
	defer s.RUnlock()
	keys := make([]string, 0, len(s.entries))
	for k, v := range s.entries {
		if s.clock.Now().After(v.expiresAt) {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}
