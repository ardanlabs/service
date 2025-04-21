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

func makeCall[T any](ctx context.Context, c *Client[T], key string, fn FetchFn[T], call *inFlightCall[T]) {
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
	call.val = response

	if c.storeMissingRecords && errors.Is(err, ErrNotFound) {
		c.StoreMissingRecord(key)
		call.err = ErrMissingRecord
		return
	}

	if err != nil && !errors.Is(err, errOnlyDistributedRecords) {
		call.err = err
		return
	}

	if errors.Is(err, errOnlyDistributedRecords) {
		call.err = onlyCachedRecords(err)
	}

	c.Set(key, response)
}

func callAndCache[T any](ctx context.Context, c *Client[T], key string, fn FetchFn[T]) (T, error) {
	c.inFlightMutex.Lock()
	if call, ok := c.inFlightMap[key]; ok {
		c.inFlightMutex.Unlock()
		call.Wait()
		return call.val, call.err
	}

	call := c.newFlight(key)
	c.inFlightMutex.Unlock()
	makeCall(ctx, c, key, fn, call)
	return call.val, call.err
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

type makeBatchCallOpts[T any] struct {
	ids   []string
	fn    BatchFetchFn[T]
	keyFn KeyFn
	call  *inFlightCall[map[string]T]
}

func makeBatchCall[T any](ctx context.Context, c *Client[T], opts makeBatchCallOpts[T]) {
	response, err := opts.fn(ctx, opts.ids)
	for id, record := range response {
		// We never want to discard values from the fetch functions, even if they
		// return an error. Instead, we'll pass them to the user along with any
		// errors and let them decide what to do.
		opts.call.val[id] = record
		// However, we'll only write them to the cache if the fetchFunction returned a non-nil error.
		if err == nil || errors.Is(err, errOnlyDistributedRecords) {
			c.Set(opts.keyFn(id), record)
		}
	}

	if err != nil && !errors.Is(err, errOnlyDistributedRecords) {
		opts.call.err = err
		return
	}

	if errors.Is(err, errOnlyDistributedRecords) {
		opts.call.err = onlyCachedRecords(err)
	}

	// Check if we should store any of these IDs as a missing record. However, we
	// don't want to do this if we only received records from the distributed
	// storage. That means that the underlying data source errored for the ID's
	// that we didn't have in our distributed storage, and we don't know wether
	// these records are missing or not.
	if c.storeMissingRecords && len(response) < len(opts.ids) && !errors.Is(err, errOnlyDistributedRecords) {
		for _, id := range opts.ids {
			if _, ok := response[id]; !ok {
				c.StoreMissingRecord(opts.keyFn(id))
			}
		}
	}
}

type callBatchOpts[T any] struct {
	ids   []string
	keyFn KeyFn
	fn    BatchFetchFn[T]
}

func callAndCacheBatch[T any](ctx context.Context, c *Client[T], opts callBatchOpts[T]) (map[string]T, error) {
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
			batchCallOpts := makeBatchCallOpts[T]{ids: uniqueIDs, fn: opts.fn, keyFn: opts.keyFn, call: call}
			makeBatchCall(ctx, c, batchCallOpts)
		}()
	}
	c.inFlightBatchMutex.Unlock()

	var err error
	response := make(map[string]T, len(opts.ids))
	for call, callIDs := range callIDs {
		call.Wait()

		// We need to iterate through the values that WE want from this call. The batch
		// could contain hundreds of IDs, but we might only want a few of them.
		for _, id := range callIDs {
			if v, ok := call.val[id]; ok {
				response[id] = v
			}
		}

		// This handles the scenario where we either don't get an error, or are
		// using the distributed storage option and are able to get some records
		// while the request to the underlying data source fails. In the latter
		// case, we'll continue to accumulate partial responses as long as the only
		// issue is cached-only records.
		if err == nil || errors.Is(call.err, ErrOnlyCachedRecords) {
			err = call.err
			continue
		}

		// For any other kind of error, we'll short‑circuit the function and return.
		return response, call.err
	}

	return response, err
}
