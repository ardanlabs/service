package sturdyc

import (
	"context"
	"errors"
)

func fetchAndCache[V, T any](ctx context.Context, c *Client[T], key string, fetchFn FetchFn[V]) (V, error) {
	response, err := fetchFn(ctx)

	if err != nil && c.storeMisses && errors.Is(err, ErrStoreMissingRecord) {
		var zero T
		c.SetMissing(key, zero, true)
		return response, ErrMissingRecord
	}

	if err != nil {
		return response, err
	}

	res, ok := any(response).(T)
	if !ok {
		return response, ErrInvalidType
	}

	c.SetMissing(key, res, false)
	return response, nil
}

func fetchAndCacheBatch[V, T any](ctx context.Context, c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[V]) (map[string]V, error) {
	response, err := fetchFn(ctx, ids)
	if err != nil {
		return response, err
	}

	// Check if we should store any of these IDs as a missing record.
	if c.storeMisses && len(response) < len(ids) {
		for _, id := range ids {
			if _, ok := response[id]; !ok {
				var zero T
				c.SetMissing(keyFn(id), zero, true)
			}
		}
	}

	// Store the records in the cache.
	for id, record := range response {
		v, ok := any(record).(T)
		if !ok {
			continue
		}
		c.SetMissing(keyFn(id), v, false)
	}

	return response, nil
}
