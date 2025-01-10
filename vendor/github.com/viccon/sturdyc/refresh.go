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
	if err != nil && !errors.Is(err, errOnlyDistributedRecords) {
		return
	}

	// Check if any of the records have been deleted at the data source.
	for _, id := range ids {
		_, okCache, _, _, _ := c.getWithState(keyFn(id))
		_, okResponse := response[id]

		if okResponse {
			continue
		}

		if !c.storeMissingRecords && !okResponse && okCache {
			c.Delete(keyFn(id))
		}

		// If we're only getting records from the distributed storage, it means that we weren't able to get
		// the remaining IDs for the batch from the underlying data source. We don't want to store these
		// as missing records because we don't know if they're missing or not.
		if c.storeMissingRecords && !okResponse && !errors.Is(err, errOnlyDistributedRecords) {
			c.StoreMissingRecord(keyFn(id))
		}
	}

	// Cache the refreshed records.
	for id, record := range response {
		c.Set(keyFn(id), record)
	}
}
