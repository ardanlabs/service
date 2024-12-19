package sturdyc

import (
	"context"
	"errors"
	"maps"
)

func (c *Client[T]) groupIDs(ids []string, keyFn KeyFn) (hits map[string]T, misses, refreshes []string) {
	hits = make(map[string]T)
	misses = make([]string, 0)
	refreshes = make([]string, 0)

	for _, id := range ids {
		key := keyFn(id)
		value, exists, markedAsMissing, shouldRefresh := c.getWithState(key)

		// Check if the record should be refreshed in the background.
		if shouldRefresh {
			refreshes = append(refreshes, id)
		}

		if markedAsMissing {
			continue
		}

		if !exists {
			misses = append(misses, id)
			continue
		}

		hits[id] = value
	}
	return hits, misses, refreshes
}

func getFetch[V, T any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (T, error) {
	wrappedFetch := wrap[T](distributedFetch(c, key, fetchFn))

	// Begin by checking if we have the item in our cache.
	value, ok, markedAsMissing, shouldRefresh := c.getWithState(key)

	if shouldRefresh {
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
// Additionally, when background refreshes are enabled, GetOrFetch determines if the record
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
	cachedRecords, cacheMisses, idsToRefresh := c.groupIDs(ids, keyFn)

	// If any records need to be refreshed, we'll do so in the background.
	if len(idsToRefresh) > 0 {
		c.safeGo(func() {
			if c.bufferRefreshes {
				bufferBatchRefresh(c, idsToRefresh, keyFn, wrappedFetch)
				return
			}
			c.refreshBatch(idsToRefresh, keyFn, wrappedFetch)
		})
	}

	// If we were able to retrieve all records from the cache, we can return them straight away.
	if len(cacheMisses) == 0 {
		return cachedRecords, nil
	}

	callBatchOpts := callBatchOpts[T, T]{ids: cacheMisses, keyFn: keyFn, fn: wrappedFetch}
	response, err := callAndCacheBatch(ctx, c, callBatchOpts)
	if err != nil && !errors.Is(err, ErrOnlyCachedRecords) {
		if len(cachedRecords) > 0 {
			return cachedRecords, ErrOnlyCachedRecords
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
