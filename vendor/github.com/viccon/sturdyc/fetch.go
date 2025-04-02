package sturdyc

import (
	"context"
	"errors"
	"maps"
)

func (c *Client[T]) groupIDs(ids []string, keyFn KeyFn) (hits map[string]T, misses, backgroundRefreshes, synchronousRefreshes []string) {
	hits = make(map[string]T)
	misses = make([]string, 0)
	backgroundRefreshes = make([]string, 0)
	synchronousRefreshes = make([]string, 0)

	for _, id := range ids {
		key := keyFn(id)
		value, exists, markedAsMissing, backgroundRefresh, synchronousRefresh := c.getWithState(key)

		if synchronousRefresh {
			synchronousRefreshes = append(synchronousRefreshes, id)
		}

		// Check if the record should be refreshed in the background.
		if backgroundRefresh && !synchronousRefresh {
			backgroundRefreshes = append(backgroundRefreshes, id)
		}

		if markedAsMissing {
			continue
		}

		// If the record should be synchronously refreshed, it's going to be added to both the hits and misses maps.
		if !exists {
			misses = append(misses, id)
			continue
		}

		hits[id] = value
	}
	return hits, misses, backgroundRefreshes, synchronousRefreshes
}

func getFetch[V, T any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (T, error) {
	value, ok, markedAsMissing, backgroundRefresh, synchronousRefresh := c.getWithState(key)
	wrappedFetch := wrap[T](distributedFetch(c, key, fetchFn))

	if synchronousRefresh {
		res, err := callAndCache(ctx, c, key, wrappedFetch)
		//  Check if the record has been deleted at the source. If it has, we'll
		//  delete it from the cache too. NOTE: The callAndCache function converts
		//  ErrNotFound to ErrMissingRecord if missing record storage is enabled.
		if ok && errors.Is(err, ErrNotFound) {
			c.Delete(key)
		}

		if errors.Is(err, ErrMissingRecord) || errors.Is(err, ErrNotFound) {
			return res, err
		}

		// If the call to synchrounously refresh the record failed,
		// we'll return the latest value if we have it in the cache
		// along with a ErrOnlyCachedRecords error. The consumer can
		// then decide whether to proceed with the cached data or to
		// propagate the error.
		if err != nil && ok {
			return value, errors.Join(ErrOnlyCachedRecords, err)
		}

		return res, err
	}

	if backgroundRefresh {
		c.safeGo(func() {
			c.refresh(key, wrappedFetch)
		})
	}

	if markedAsMissing {
		return value, ErrMissingRecord
	}

	if ok {
		return value, nil
	}

	return callAndCache(ctx, c, key, wrappedFetch)
}

// GetOrFetch attempts to retrieve the specified key from the cache. If the value
// is absent, it invokes the fetchFn function to obtain it and then stores the result.
// Additionally, when early refreshes are enabled, GetOrFetch determines if the record
// needs refreshing and, if necessary, schedules this task for background execution.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	key - The key to be fetched.
//	fetchFn - Used to retrieve the data from the underlying data source if the key is not found in the cache.
//
// Returns:
//
//	The value corresponding to the key and an error if one occurred.
func (c *Client[T]) GetOrFetch(ctx context.Context, key string, fetchFn FetchFn[T]) (T, error) {
	return getFetch[T, T](ctx, c, key, fetchFn)
}

// GetOrFetch is a convenience function that performs type assertion on the result of client.GetOrFetch.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	c - The cache client.
//	key - The key to be fetched.
//	fetchFn - Used to retrieve the data from the underlying data source if the key is not found in the cache.
//
// Returns:
//
//	The value corresponding to the key and an error if one occurred.
//
// Type Parameters:
//
//	V - The type returned by the fetchFn. Must be assignable to T.
//	T - The type stored in the cache.
func GetOrFetch[V, T any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (V, error) {
	res, err := getFetch[V, T](ctx, c, key, fetchFn)
	return unwrap[V](res, err)
}

func getFetchBatch[V, T any](ctx context.Context, c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[V]) (map[string]T, error) {
	wrappedFetch := wrapBatch[T](distributedBatchFetch[V, T](c, keyFn, fetchFn))
	cachedRecords, cacheMisses, idsToBackgroundRefresh, idsToSynchronouslyRefresh := c.groupIDs(ids, keyFn)

	// Schedule background refreshes.
	if len(idsToBackgroundRefresh) > 0 {
		c.safeGo(func() {
			if c.bufferRefreshes {
				bufferBatchRefresh(c, idsToBackgroundRefresh, keyFn, wrappedFetch)
				return
			}
			c.refreshBatch(idsToBackgroundRefresh, keyFn, wrappedFetch)
		})
	}

	// If we were able to retrieve all records from the cache, we can return them straight away.
	if len(cacheMisses) == 0 && len(idsToSynchronouslyRefresh) == 0 {
		return cachedRecords, nil
	}

	// Create a list of the IDs that we're going to fetch from the underlying data source or distributed storage.
	cacheMissesAndSyncRefreshes := make([]string, 0, len(cacheMisses)+len(idsToSynchronouslyRefresh))
	cacheMissesAndSyncRefreshes = append(cacheMissesAndSyncRefreshes, cacheMisses...)
	cacheMissesAndSyncRefreshes = append(cacheMissesAndSyncRefreshes, idsToSynchronouslyRefresh...)

	callBatchOpts := callBatchOpts[T, T]{ids: cacheMissesAndSyncRefreshes, keyFn: keyFn, fn: wrappedFetch}
	response, err := callAndCacheBatch(ctx, c, callBatchOpts)

	// If we did a call to synchronously refresh some of the records, and it
	// didn't fail, we'll have to check if any of the IDs have been deleted at
	// the underlying data source. If they have, we'll have to delete them from
	// the cache and remove them from the cachedRecords map so that we don't
	// return them.
	if err == nil && len(idsToSynchronouslyRefresh) > 0 {
		for _, id := range idsToSynchronouslyRefresh {
			// If we have it in the cache, but not in the response, it means
			// that the ID no longer exists at the underlying data source.
			_, okResponse := response[id]
			_, okCache := cachedRecords[id]
			if okCache && !okResponse {
				if !c.storeMissingRecords {
					c.Delete(keyFn(id))
				}
				delete(cachedRecords, id)
			}
		}
	}

	if err != nil && !errors.Is(err, ErrOnlyCachedRecords) {
		if len(cachedRecords) > 0 {
			return cachedRecords, errors.Join(ErrOnlyCachedRecords, err)
		}
		return cachedRecords, err
	}

	maps.Copy(cachedRecords, response)
	return cachedRecords, err
}

// GetOrFetchBatch attempts to retrieve the specified ids from the cache. If
// any of the values are absent, it invokes the fetchFn function to obtain them
// and then stores the result. Additionally, when background refreshes are
// enabled, GetOrFetch determines if any of the records need refreshing and, if
// necessary, schedules this to be performed in the background.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	ids - The list of IDs to be fetched.
//	keyFn - Used to generate the cache key for each ID.
//	fetchFn - Used to retrieve the data from the underlying data source if any IDs are not found in the cache.
//
// Returns:
//
//	A map of IDs to their corresponding values and an error if one occurred.
func (c *Client[T]) GetOrFetchBatch(ctx context.Context, ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) (map[string]T, error) {
	return getFetchBatch[T, T](ctx, c, ids, keyFn, fetchFn)
}

// GetOrFetchBatch is a convenience function that performs type assertion on the
// result of client.GetOrFetchBatch.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	c - The cache client.
//	ids - The list of IDs to be fetched.
//	keyFn - Used to prefix each ID in order to create a unique cache key.
//	fetchFn - Used to retrieve the data from the underlying data source.
//
// Returns:
//
//	A map of ids to their corresponding values and an error if one occurred.
//
// Type Parameters:
//
//	V - The type returned by the fetchFn. Must be assignable to T.
//	T - The type stored in the cache.

func GetOrFetchBatch[V, T any](ctx context.Context, c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[V]) (map[string]V, error) {
	res, err := getFetchBatch[V, T](ctx, c, ids, keyFn, fetchFn)
	return unwrapBatch[V](res, err)
}
