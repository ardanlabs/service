package sturdyc

import (
	"context"
	"errors"
)

func (c *Client[T]) refresh(key string, fetchFn FetchFn[T]) {
	response, err := fetchFn(context.Background())
	if err != nil {
		if c.storeMissingRecords && errors.Is(err, ErrNotFound) {
			c.StoreMissingRecord(key)
		}
		if !c.storeMissingRecords && errors.Is(err, ErrNotFound) {
			c.Delete(key)
		}
		return
	}
	c.Set(key, response)
}

func (c *Client[T]) refreshBatch(ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) {
	c.reportBatchRefreshSize(len(ids))
	response, err := fetchFn(context.Background(), ids)
	if err != nil {
		return
	}

	// Check if any of the records have been deleted at the data source.
	for _, id := range ids {
		_, okCache, _, _ := c.get(keyFn(id))
		_, okResponse := response[id]

		if okResponse {
			continue
		}

		if !c.storeMissingRecords && !okResponse && okCache {
			c.Delete(keyFn(id))
		}

		if c.storeMissingRecords && !okResponse {
			c.StoreMissingRecord(keyFn(id))
		}
	}

	// Cache the refreshed records.
	for id, record := range response {
		c.Set(keyFn(id), record)
	}
}
