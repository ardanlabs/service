package sturdyc

import "time"

type Option func(*Config)

// WithMetrics is used to make the cache report metrics.
func WithMetrics(recorder MetricsRecorder) Option {
	return func(c *Config) {
		recorder.ObserveCacheSize(c.getSize)
		c.metricsRecorder = &distributedMetricsRecorder{recorder}
	}
}

// WithClock can be used to change the clock that the cache uses. This is useful for testing.
func WithClock(clock Clock) Option {
	return func(c *Config) {
		c.clock = clock
	}
}

// WithEvictionInterval sets the interval at which the cache scans a shard to
// evict expired entries. Setting this to a higher value will increase cache
// performance and is advised if you don't think you'll exceed the capacity.
// If the capacity is reached, the cache will still trigger an eviction.
func WithEvictionInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.evictionInterval = interval
	}
}

// WithNoContinuousEvictions improves cache performance when the cache capacity
// is unlikely to be exceeded. While this setting disables the continuous
// eviction job, it still allows for the eviction of the least recently used
// items once the cache reaches its full capacity.
func WithNoContinuousEvictions() Option {
	return func(c *Config) {
		c.disableContinuousEvictions = true
	}
}

// WithMissingRecordStorage allows the cache to mark keys as missing from the
// underlying data source. This allows you to stop streams of outgoing requests
// for requests that don't exist. The keys will still have the same TTL and
// refresh durations as any of the other record in the cache.
func WithMissingRecordStorage() Option {
	return func(c *Config) {
		c.storeMissingRecords = true
	}
}

// WithEarlyRefreshes instructs the cache to refresh the keys that are in
// active rotation, thereby preventing them from ever expiring. This can have a
// significant impact on your application's latency as you're able to
// continuously serve frequently used keys from memory. An asynchronous
// background refresh gets scheduled when a key is requested again after a
// random time between minRefreshTime and maxRefreshTime has passed. This is an
// important distinction because it means that the cache won't just naively
// refresh every key it's ever seen. The third argument to this function will
// also allow you to provide a duration for when a refresh should become
// synchronous. If any of the refreshes were to fail, you'll get the latest
// data from the cache for the duration of the TTL.
func WithEarlyRefreshes(minAsyncRefreshTime, maxAsyncRefreshTime, syncRefreshTime, retryBaseDelay time.Duration) Option {
	return func(c *Config) {
		c.earlyRefreshes = true
		c.minAsyncRefreshTime = minAsyncRefreshTime
		c.maxAsyncRefreshTime = maxAsyncRefreshTime
		c.syncRefreshTime = syncRefreshTime
		c.retryBaseDelay = retryBaseDelay
	}
}

// WithRefreshCoalescing will make the cache refresh data from batchable
// endpoints more efficiently. It is going to create a buffer for each cache
// key permutation, and gather IDs until the bufferSize is reached, or the
// bufferDuration has passed.
//
// NOTE: This requires the WithEarlyRefreshes functionality to be enabled.
func WithRefreshCoalescing(bufferSize int, bufferDuration time.Duration) Option {
	return func(c *Config) {
		c.bufferRefreshes = true
		c.bufferSize = bufferSize
		c.bufferTimeout = bufferDuration
		c.permutationBufferMap = make(map[string]*buffer)
	}
}

// WithRelativeTimeKeyFormat allows you to control the truncation of time.Time
// values that are being passed in to the cache key functions.
func WithRelativeTimeKeyFormat(truncation time.Duration) Option {
	return func(c *Config) {
		c.useRelativeTimeKeyFormat = true
		c.keyTruncation = truncation
	}
}

// WithLog allows you to set a custom logger for the cache. The cache isn't chatty,
// and will only log warnings and errors that would be a nightmare to debug. If you
// absolutely don't want any logs, you can pass in the sturdyc.NoopLogger.
func WithLog(log Logger) Option {
	return func(c *Config) {
		c.log = log
	}
}

// WithDistributedStorage allows you to use the cache with a distributed
// key-value store. The "GetOrFetch" and "GetOrFetchBatch" functions will check
// this store first and only proceed to the underlying data source if the key
// is missing. When a record is retrieved from the underlying data source, it
// is written both to memory and to the distributed storage. You are
// responsible for setting TTL and eviction policies for the distributed
// storage. Sturdyc will only read and write records.
func WithDistributedStorage(storage DistributedStorage) Option {
	return func(c *Config) {
		c.distributedStorage = &distributedStorage{storage}
		c.distributedEarlyRefreshes = false
	}
}

// WithDistributedStorageEarlyRefreshes is the distributed equivalent of the
// "WithEarlyRefreshes" option. It allows distributed records to be refreshed
// before their TTL expires. If a refresh fails, the cache will fall back to
// what was returned by the distributed storage. This ensures that data can be
// served for the duration of the TTL even if an upstream system goes down. To
// use this functionality, you need to implement an interface with two
// additional methods for deleting records compared to the simpler
// "WithDistributedStorage" option. This is because a distributed cache that is
// used with this option might have low refresh durations but high TTLs. If a
// record is deleted from the underlying data source, it needs to be propagated
// to the distributed storage before the TTL expires. However, please note that
// you are still responsible for managing the TTL and eviction policies for the
// distributed storage. Sturdyc will only delete records that have been removed
// at the underlying data source.
func WithDistributedStorageEarlyRefreshes(storage DistributedStorageWithDeletions, refreshAfter time.Duration) Option {
	return func(c *Config) {
		c.distributedStorage = storage
		c.distributedEarlyRefreshes = true
		c.distributedRefreshAfterDuration = refreshAfter
	}
}

// WithDistributedMetrics instructs the cache to report additional metrics
// regarding its interaction with the distributed storage.
func WithDistributedMetrics(metricsRecorder DistributedMetricsRecorder) Option {
	return func(c *Config) {
		metricsRecorder.ObserveCacheSize(c.getSize)
		c.metricsRecorder = metricsRecorder
	}
}

// validateConfig is a helper function that panics if the cache has been configured incorrectly.
func validateConfig(capacity, numShards int, ttl time.Duration, evictionPercentage int, cfg *Config) {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	if numShards <= 0 {
		panic("numShards must be greater than 0")
	}

	if ttl <= 0 {
		panic("ttl must be greater than 0")
	}

	if evictionPercentage < 0 || evictionPercentage > 100 {
		panic("evictionPercentage must be between 0 and 100")
	}

	if !cfg.earlyRefreshes && cfg.bufferRefreshes {
		panic("refresh buffering requires early refreshes to be enabled")
	}

	if cfg.bufferRefreshes && cfg.bufferSize < 1 {
		panic("batchSize must be greater than 0")
	}

	if cfg.bufferRefreshes && cfg.bufferTimeout < 1 {
		panic("bufferTimeout must be greater than 0")
	}

	if cfg.evictionInterval < 1 {
		panic("evictionInterval must be greater than 0")
	}

	if cfg.minAsyncRefreshTime > cfg.maxAsyncRefreshTime {
		panic("minRefreshTime must be less than or equal to maxRefreshTime")
	}

	if cfg.maxAsyncRefreshTime > cfg.syncRefreshTime {
		panic("maxRefreshTime must be less than or equal to synchronousRefreshTime")
	}

	if cfg.retryBaseDelay < 0 {
		panic("retryBaseDelay must be greater than or equal to 0")
	}
}
