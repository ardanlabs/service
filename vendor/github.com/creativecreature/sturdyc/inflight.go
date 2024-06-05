package sturdyc

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type inFlightCall[T any] struct {
	sync.WaitGroup
	val T
	err error
}

// newFlight should be called with a lock.
func (c *Client[T]) newFlight(key string) *inFlightCall[T] {
	call := new(inFlightCall[T])
	call.Add(1)
	c.inFlightMap[key] = call
	return call
}

func makeCall[T, V any](ctx context.Context, c *Client[T], key string, fn FetchFn[V], call *inFlightCall[T]) {
	defer func() {
		if err := recover(); err != nil {
			call.err = fmt.Errorf("sturdyc: panic recovered: %v", err)
		}
		call.Done()
		c.inFlightMutex.Lock()
		delete(c.inFlightMap, key)
		c.inFlightMutex.Unlock()
	}()

	response, err := fn(ctx)
	if err != nil && c.storeMissingRecords && errors.Is(err, ErrStoreMissingRecord) {
		c.StoreMissingRecord(key)
		call.err = ErrMissingRecord
		return
	}

	if err != nil {
		call.err = err
		return
	}

	res, ok := any(response).(T)
	if !ok {
		call.err = ErrInvalidType
		return
	}

	call.err = nil
	call.val = res
	c.Set(key, res)
}

func callAndCache[V, T any](ctx context.Context, c *Client[T], key string, fn FetchFn[V]) (V, error) {
	c.inFlightMutex.Lock()
	if call, ok := c.inFlightMap[key]; ok {
		c.inFlightMutex.Unlock()
		call.Wait()
		return unwrap[V, T](call.val, call.err)
	}

	call := c.newFlight(key)
	c.inFlightMutex.Unlock()
	makeCall(ctx, c, key, fn, call)
	return unwrap[V, T](call.val, call.err)
}

// newBatchFlight should be called with a lock.
func (c *Client[T]) newBatchFlight(ids []string, keyFn KeyFn) *inFlightCall[map[string]T] {
	call := new(inFlightCall[map[string]T])
	call.val = make(map[string]T, len(ids))
	call.Add(1)
	for _, id := range ids {
		c.inFlightBatchMap[keyFn(id)] = call
	}
	return call
}

func (c *Client[T]) endBatchFlight(ids []string, keyFn KeyFn, call *inFlightCall[map[string]T]) {
	call.Done()
	c.inFlightBatchMutex.Lock()
	for _, id := range ids {
		delete(c.inFlightBatchMap, keyFn(id))
	}
	c.inFlightBatchMutex.Unlock()
}

type makeBatchCallOpts[T, V any] struct {
	ids   []string
	fn    BatchFetchFn[V]
	keyFn KeyFn
	call  *inFlightCall[map[string]T]
}

func makeBatchCall[T, V any](ctx context.Context, c *Client[T], opts makeBatchCallOpts[T, V]) {
	response, err := opts.fn(ctx, opts.ids)
	if err != nil {
		opts.call.err = err
		return
	}

	// Check if we should store any of these IDs as a missing record.
	if c.storeMissingRecords && len(response) < len(opts.ids) {
		for _, id := range opts.ids {
			if _, ok := response[id]; !ok {
				c.StoreMissingRecord(opts.keyFn(id))
			}
		}
	}

	// Store the records in the cache.
	for id, record := range response {
		v, ok := any(record).(T)
		if !ok {
			c.log.Error("sturdyc: invalid type for ID:" + id)
			continue
		}
		c.Set(opts.keyFn(id), v)
		opts.call.val[id] = v
	}
}

type callBatchOpts[T, V any] struct {
	ids   []string
	keyFn KeyFn
	fn    BatchFetchFn[V]
}

func callAndCacheBatch[V, T any](ctx context.Context, c *Client[T], opts callBatchOpts[T, V]) (map[string]V, error) {
	c.inFlightBatchMutex.Lock()

	callIDs := make(map[*inFlightCall[map[string]T]][]string)
	uniqueIDs := make([]string, 0, len(opts.ids))
	for _, id := range opts.ids {
		if call, ok := c.inFlightBatchMap[opts.keyFn(id)]; ok {
			callIDs[call] = append(callIDs[call], id)
			continue
		}
		uniqueIDs = append(uniqueIDs, id)
	}

	if len(uniqueIDs) > 0 {
		call := c.newBatchFlight(uniqueIDs, opts.keyFn)
		callIDs[call] = append(callIDs[call], uniqueIDs...)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					call.err = fmt.Errorf("sturdyc: panic recovered: %v", err)
				}
				c.endBatchFlight(uniqueIDs, opts.keyFn, call)
			}()
			batchCallOpts := makeBatchCallOpts[T, V]{
				ids:   uniqueIDs,
				fn:    opts.fn,
				keyFn: opts.keyFn,
				call:  call,
			}
			makeBatchCall(ctx, c, batchCallOpts)
		}()
	}
	c.inFlightBatchMutex.Unlock()

	response := make(map[string]V, len(opts.ids))
	for call, callIDs := range callIDs {
		call.Wait()
		if call.err != nil {
			return response, call.err
		}

		// Remember: we need to iterate through the values we have for this call.
		// We might just need one ID and the batch could contain a hundred.
		for _, id := range callIDs {
			v, ok := call.val[id]
			if !ok {
				continue
			}

			if val, ok := any(v).(V); ok {
				response[id] = val
				continue
			}
			return response, ErrInvalidType
		}
	}

	return response, nil
}
