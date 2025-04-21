package sturdyc

import (
	"context"
)

// Passthrough attempts to retrieve the latest data by calling the provided fetchFn.
// If fetchFn encounters an error, the cache is used as a fallback.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	key - The key to be fetched.
//	fetchFn - Used to retrieve the data from the underlying data source.
//
// Returns:
//
//	The value and an error if one occurred and the key was not found in the cache.
func (c *Client[T]) Passthrough(ctx context.Context, key string, fetchFn FetchFn[T]) (T, error) {
	res, err := callAndCache(ctx, c, key, fetchFn)
	if err == nil {
		return res, nil
	}

	if value, ok := c.Get(key); ok {
		return value, nil
	}

	return res, err
}

// Passthrough is a convenience function that performs type assertion on the
// result of client.PassthroughBatch.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	c - The cache client.
//	key - The key to be fetched.
//	fetchFn - Used to retrieve the data from the underlying data source.
//
// Returns:
//
//	The value and an error if one occurred and the key was not found in the cache.
//
// Type Parameters:
//
//	V - The type returned by the fetchFn. Must be assignable to T.
//	T - The type stored in the cache.
func Passthrough[T, V any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (V, error) {
	value, err := c.Passthrough(ctx, key, wrap[T](fetchFn))
	return unwrap[V](value, err)
}

// PassthroughBatch attempts to retrieve the latest data by calling the provided fetchFn.
// If fetchFn encounters an error, the cache is used as a fallback.
//
// Parameters:
//
//	ctx - The context to be used for the request.
//	ids - The list of IDs to be fetched.
//	keyFn - Used to prefix each ID in order to create a unique cache key.
//	fetchFn - Used to retrieve the data from the underlying data source.
//
// Returns:
//
//	A map of IDs to their corresponding values, and an error if one occurred and
//	none of the IDs were found in the cache.
func (c *Client[T]) PassthroughBatch(ctx context.Context, ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) (map[string]T, error) {
	res, err := callAndCacheBatch(ctx, c, callBatchOpts[T]{ids, keyFn, fetchFn})
	if err == nil {
		return res, nil
	}

	values := c.GetManyKeyFn(ids, keyFn)
	if len(values) > 0 {
		return values, nil
	}

	return res, err
}

// PassthroughBatch is a convenience function that performs type assertion on the
// result of client.PassthroughBatch.
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
func PassthroughBatch[V, T any](ctx context.Context, c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[V]) (map[string]V, error) {
	res, err := c.PassthroughBatch(ctx, ids, keyFn, wrapBatch[T](fetchFn))
	return unwrapBatch[V](res, err)
}
