package sturdyc

import (
	"context"
	"errors"
)

func (c *Client[T]) refresh(key string, fetchFn FetchFn[T]) {
	response, err := fetchFn(context.Background())
	if err != nil {
		if c.storeMisses && errors.Is(err, ErrStoreMissingRecord) {
			c.SetMissing(key, response, true)
		}
		if errors.Is(err, ErrDeleteRecord) {
			c.Delete(key)
		}
		return
	}
	c.SetMissing(key, response, false)
}

func (c *Client[T]) refreshBatch(ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) {
	if c.metricsRecorder != nil {
		c.metricsRecorder.CacheBatchRefreshSize(len(ids))
	}

	response, err := fetchFn(context.Background(), ids)
	if err != nil {
		return
	}

	// Check if any of the records have been deleted at the data source.
	for _, id := range ids {
		_, okCache, _, _ := c.get(keyFn(id))
		v, okResponse := response[id]

		if okResponse {
			continue
		}

		if !c.storeMisses && !okResponse && okCache {
			c.Delete(keyFn(id))
		}

		if c.storeMisses && !okResponse {
			c.SetMissing(keyFn(id), v, true)
		}
	}

	// Cache the refreshed records.
	for id, record := range response {
		c.SetMissing(keyFn(id), record, false)
	}
}
