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

// WithBackgroundRefreshes...
func WithBackgroundRefreshes(minRefreshTime, maxRefreshTime, retryBaseDelay time.Duration) Option {
	return func(c *Config) {
		c.refreshInBackground = true
		c.minRefreshTime = minRefreshTime
		c.maxRefreshTime = maxRefreshTime
		c.retryBaseDelay = retryBaseDelay
	}
}

// WithMissingRecordStorage allows the cache to mark keys as missing from the underlying data source.
func WithMissingRecordStorage() Option {
	return func(c *Config) {
		c.storeMissingRecords = true
	}
}

// WithRefreshBuffering will make the cache refresh data from batchable
// endpoints more efficiently. It is going to create a buffer for each cache
// key permutation, and gather IDs until the bufferSize is reached, or the
// bufferDuration has passed.
func WithRefreshBuffering(bufferSize int, bufferDuration time.Duration) Option {
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
// absolutely don't want any logs, you can pass in the sturydc.NoopLogger.
func WithLog(log Logger) Option {
	return func(c *Config) {
		c.log = log
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

	if !cfg.refreshInBackground && cfg.bufferRefreshes {
		panic("refresh buffering requires background refreshes to be enabled")
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

	if cfg.minRefreshTime > cfg.maxRefreshTime {
		panic("minRefreshTime must be less than or equal to maxRefreshTime")
	}

	if cfg.retryBaseDelay < 0 {
		panic("retryBaseDelay must be greater than or equal to 0")
	}
}
