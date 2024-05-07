package sturdyc

import (
	"context"
	"sync"
	"time"

	"github.com/cespare/xxhash"
)

type MetricsRecorder interface {
	CacheHit()
	CacheMiss()
	Eviction()
	ForcedEviction()
	EntriesEvicted(int)
	ShardIndex(int)
	CacheBatchRefreshSize(size int)
	ObserveCacheSize(callback func() int)
}

// FetchFn Fetch represents a function that can be used to fetch a single record from a data source.
type FetchFn[T any] func(ctx context.Context) (T, error)

// BatchFetchFn represents a function that can be used to fetch multiple records from a data source.
type BatchFetchFn[T any] func(ctx context.Context, ids []string) (map[string]T, error)

type BatchResponse[T any] map[string]T

// KeyFn is called invoked for each record that a batch fetch
// operation returns. It is used to create unique cache keys.
type KeyFn func(id string) string

// Config represents the configuration that can be applied to the cache.
type Config struct {
	clock            Clock
	evictionInterval time.Duration
	metricsRecorder  MetricsRecorder

	refreshesEnabled bool
	minRefreshTime   time.Duration
	maxRefreshTime   time.Duration
	retryBaseDelay   time.Duration
	storeMisses      bool

	bufferRefreshes       bool
	batchMutex            sync.Mutex
	batchSize             int
	bufferTimeout         time.Duration
	bufferPermutationIDs  map[string][]string
	bufferPermutationChan map[string]chan<- []string

	passthroughPercentage int
	passthroughBuffering  bool

	useRelativeTimeKeyFormat bool
	keyTruncation            time.Duration
	getSize                  func() int
}

// Client represents a cache client that can be used to store and retrieve values.
type Client[T any] struct {
	*Config
	shards    []*shard[T]
	nextShard int
}

// New creates a new Client instance with the specified configuration.
//
// `capacity` defines the maximum number of entries that the cache can store.
// `numShards` Is used to set the number of shards. Has to be greater than 0.
// `ttl` Sets the time to live for each entry in the cache. Has to be greater than 0.
// `evictionPercentage` Percentage of items to evict when the cache exceeds its capacity.
// `opts` allows for additional configurations to be applied to the cache client.
func New[T any](capacity, numShards int, ttl time.Duration, evictionPercentage int, opts ...Option) *Client[T] {
	// Create an emptu client and setup the default configuration.
	client := &Client[T]{}
	cfg := &Config{
		clock:                 NewClock(),
		passthroughPercentage: 100,
		evictionInterval:      ttl / time.Duration(numShards),
		getSize:               client.Size,
	}

	// Apply the options to the configuration.
	client.Config = cfg
	for _, opt := range opts {
		opt(cfg)
	}

	validateConfig(capacity, numShards, ttl, evictionPercentage, cfg)

	// Next, we'll create the shards. It is important that
	// we do this after we've applited the options.
	shardSize := capacity / numShards
	shards := make([]*shard[T], numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = newShard[T](shardSize, ttl, evictionPercentage, cfg)
	}
	client.shards = shards
	client.nextShard = 0

	// Run evictions on the shards in a separate goroutine.
	client.startEvictions()

	return client
}

// startEvictions is going to be running in a separate goroutine that we're going to prevent from ever exiting.
func (c *Client[T]) startEvictions() {
	go func() {
		ticker, stop := c.clock.NewTicker(c.evictionInterval)
		defer stop()
		for range ticker {
			if c.metricsRecorder != nil {
				c.metricsRecorder.Eviction()
			}
			c.shards[c.nextShard].evictExpired()
			c.nextShard = (c.nextShard + 1) % len(c.shards)
		}
	}()
}

// getShard returns the shard that should be used for the specified key.
func (c *Client[T]) getShard(key string) *shard[T] {
	hash := xxhash.Sum64String(key)
	shardIndex := hash % uint64(len(c.shards))
	if c.metricsRecorder != nil {
		c.metricsRecorder.ShardIndex(int(shardIndex))
	}
	return c.shards[shardIndex]
}

// reportCacheHits is used to report cache hits and misses to the metrics recorder.
func (c *Client[T]) reportCacheHits(cacheHit bool) {
	if c.metricsRecorder == nil {
		return
	}
	if !cacheHit {
		c.metricsRecorder.CacheMiss()
		return
	}
	c.metricsRecorder.CacheHit()
}

func (c *Client[T]) get(key string) (value T, exists, ignore, refresh bool) {
	shard := c.getShard(key)
	val, exists, ignore, refresh := shard.get(key)
	c.reportCacheHits(exists)
	return val, exists, ignore, refresh
}

func (c *Client[T]) Get(key string) (T, bool) {
	shard := c.getShard(key)
	val, ok, ignore, _ := shard.get(key)
	c.reportCacheHits(ok)
	return val, ok && !ignore
}

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

// Size returns the number of entries in the cache.
func (c *Client[T]) Size() int {
	var sum int
	for _, shard := range c.shards {
		sum += shard.size()
	}
	return sum
}

// Delete removes a single entry from the cache.
func (c *Client[T]) Delete(key string) {
	shard := c.getShard(key)
	shard.delete(key)
}
