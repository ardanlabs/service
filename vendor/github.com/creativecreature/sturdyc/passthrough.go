package sturdyc

import (
	"context"
)

// Passthrough is always going to try and retrieve the latest data by calling the
// fetchFn. The cache is used as a fallback if the fetchFn returns an error.
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

// Passthrough is a convenience function that performs type assertion on the result of client.Passthrough.
func Passthrough[T, V any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (V, error) {
	value, err := c.Passthrough(ctx, key, wrap[T](fetchFn))
	return unwrap[V](value, err)
}

// PassthroughBatch is always going to try and retrieve the latest data by calling
// the fetchFn. The cache is used as a fallback if the fetchFn returns an error.
func (c *Client[T]) PassthroughBatch(ctx context.Context, ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) (map[string]T, error) {
	res, err := callAndCacheBatch(ctx, c, callBatchOpts[T, T]{ids, keyFn, fetchFn})
	if err == nil {
		return res, nil
	}

	values := c.GetManyKeyFn(ids, keyFn)
	if len(values) > 0 {
		return values, nil
	}

	return res, err
}

// Passthrough is a convenience function that performs type assertion on the result of client.PassthroughBatch.
func PassthroughBatch[V, T any](ctx context.Context, c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[V]) (map[string]V, error) {
	res, err := c.PassthroughBatch(ctx, ids, keyFn, wrapBatch[T](fetchFn))
	return unwrapBatch[V](res, err)
}
