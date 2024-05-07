package sturdyc

import "time"

type Option func(*Config)

// WithMetrics is used to make the cache report metrics.
func WithMetrics(recorder MetricsRecorder) Option {
	return func(c *Config) {
		recorder.ObserveCacheSize(c.getSize)
		c.metricsRecorder = recorder
	}
}

// WithClock can be used to change the clock that the cache uses. This is useful for testing.
func WithClock(clock Clock) Option {
	return func(c *Config) {
		c.clock = clock
	}
}

// WithEvictionInterval sets the interval at which the cache scans a shard to evict expired entries.
func WithEvictionInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.evictionInterval = interval
	}
}

// WithStampedeProtection makes the cache shield the underlying data sources from
// cache stampedes. Cache stampedes occur when many requests for a particular piece
// of data (which has just expired or been evicted from the cache) come in at once.
// This can cause all requests to fetch the data concurrently, which may result in
// a significant burst of outgoing requests to the underlying data source.
func WithStampedeProtection(minRefreshTime, maxRefreshTime, retryBaseDelay time.Duration, storeMisses bool) Option {
	return func(c *Config) {
		c.refreshesEnabled = true
		c.minRefreshTime = minRefreshTime
		c.maxRefreshTime = maxRefreshTime
		c.retryBaseDelay = retryBaseDelay
		c.storeMisses = storeMisses
	}
}

// WithRefreshBuffering will make the cache refresh data from batchable
// endpoints more efficiently. It is going to create a buffer for each cache
// key permutation, and gather IDs until the "batchSize" is reached, or the
// "maxBufferTime" has passed.
func WithRefreshBuffering(batchSize int, maxBufferTime time.Duration) Option {
	return func(c *Config) {
		c.bufferRefreshes = true
		c.batchSize = batchSize
		c.bufferTimeout = maxBufferTime
		c.bufferPermutationIDs = make(map[string][]string)
		c.bufferPermutationChan = make(map[string]chan<- []string)
	}
}

// WithPassthroughPercentage controls the rate at which requests are allowed through
// by the passthrough caching functions. For example, setting the percentage parameter
// to 50 would allow half of the requests to through.
func WithPassthroughPercentage(percentage int) Option {
	return func(c *Config) {
		c.passthroughPercentage = percentage
	}
}

// WithPassthroughBuffering allows you to decide if the batchable passthrough
// requests should be buffered and batched more efficiently.
func WithPassthroughBuffering() Option {
	return func(c *Config) {
		c.passthroughBuffering = true
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

	if !cfg.refreshesEnabled && cfg.bufferRefreshes {
		panic("refresh buffering requires stampede protection to be enabled")
	}

	if cfg.bufferRefreshes && cfg.batchSize < 1 {
		panic("batchSize must be greater than 0")
	}

	if cfg.bufferRefreshes && cfg.bufferTimeout < 1 {
		panic("bufferTimeout must be greater than 0")
	}

	if cfg.evictionInterval < 1 {
		panic("evictionInterval must be greater than 0")
	}

	if cfg.minRefreshTime > cfg.maxRefreshTime {
		panic("minRefreshTime must be less than or equal to maxRefreshTime")
	}

	if cfg.retryBaseDelay < 0 {
		panic("retryBaseDelay must be greater than or equal to 0")
	}

	if cfg.passthroughPercentage < 1 || cfg.passthroughPercentage > 100 {
		panic("passthroughPercentage must be between 1 and 100")
	}
}
