package sturdyc

import (
	"context"
	"errors"
	"fmt"
)

// safeGo is a helper that prevents panics in any of the goroutines
// that are running in the background from crashing the process.
func (c *Client[T]) safeGo(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c.log.Error(fmt.Sprintf("sturdyc: panic recovered: %v", err))
			}
		}()
		fn()
	}()
}

func wrap[T, V any](fetchFn FetchFn[V]) FetchFn[T] {
	return func(ctx context.Context) (T, error) {
		res, err := fetchFn(ctx)
		if err != nil {
			var zero T
			return zero, err
		}
		val, ok := any(res).(T)
		if !ok {
			var zero T
			return zero, ErrInvalidType
		}
		return val, nil
	}
}

func unwrap[V, T any](val T, err error) (V, error) {
	v, ok := any(val).(V)
	if !ok {
		return v, ErrInvalidType
	}

	return v, err
}

func wrapBatch[T, V any](fetchFn BatchFetchFn[V]) BatchFetchFn[T] {
	return func(ctx context.Context, ids []string) (map[string]T, error) {
		resV, err := fetchFn(ctx, ids)
		if err != nil && !errors.Is(err, errOnlyDistributedRecords) {
			return map[string]T{}, err
		}

		resT := make(map[string]T, len(resV))
		for id, v := range resV {
			val, ok := any(v).(T)
			if !ok {
				return resT, ErrInvalidType
			}
			resT[id] = val
		}

		return resT, err
	}
}

func unwrapBatch[V, T any](values map[string]T, err error) (map[string]V, error) {
	vals := make(map[string]V, len(values))
	for id, v := range values {
		val, ok := any(v).(V)
		if !ok {
			return vals, ErrInvalidType
		}
		vals[id] = val
	}
	return vals, err
}
