package sturdyc

// SetMissing writes a single value to the cache. Returns true if it triggered an eviction.
func (c *Client[T]) SetMissing(key string, value T, isMissingRecord bool) bool {
	shard := c.getShard(key)
	return shard.set(key, value, isMissingRecord)
}

// Set writes a single value to the cache. Returns true if it triggered an eviction.
func (c *Client[T]) Set(key string, value T) bool {
	return c.SetMissing(key, value, false)
}

// SetMany writes multiple values to the cache. Returns true if it triggered an eviction.
func (c *Client[T]) SetMany(records map[string]T, cacheKeyFn KeyFn) bool {
	var triggeredEviction bool
	for id, value := range records {
		evicted := c.SetMissing(cacheKeyFn(id), value, false)
		if evicted {
			triggeredEviction = true
		}
	}
	return triggeredEviction
}
