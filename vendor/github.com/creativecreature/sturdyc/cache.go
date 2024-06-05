package sturdyc

import (
	"context"
	"log/slog"
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
	log              Logger

	refreshInBackground bool
	minRefreshTime      time.Duration
	maxRefreshTime      time.Duration
	retryBaseDelay      time.Duration
	storeMissingRecords bool

	bufferRefreshes      bool
	batchMutex           sync.Mutex
	bufferSize           int
	bufferTimeout        time.Duration
	permutationBufferMap map[string]*buffer

	useRelativeTimeKeyFormat bool
	keyTruncation            time.Duration
	getSize                  func() int
}

// Client represents a cache client that can be used to store and retrieve values.
type Client[T any] struct {
	*Config
	shards             []*shard[T]
	nextShard          int
	inFlightMutex      sync.Mutex
	inFlightBatchMutex sync.Mutex
	inFlightMap        map[string]*inFlightCall[T]
	inFlightBatchMap   map[string]*inFlightCall[map[string]T]
}

// New creates a new Client instance with the specified configuration.
//
// `capacity` defines the maximum number of entries that the cache can store.
// `numShards` Is used to set the number of shards. Has to be greater than 0.
// `ttl` Sets the time to live for each entry in the cache. Has to be greater than 0.
// `evictionPercentage` Percentage of items to evict when the cache exceeds its capacity.
// `opts` allows for additional configurations to be applied to the cache client.
func New[T any](capacity, numShards int, ttl time.Duration, evictionPercentage int, opts ...Option) *Client[T] {
	client := &Client[T]{
		inFlightMap:      make(map[string]*inFlightCall[T]),
		inFlightBatchMap: make(map[string]*inFlightCall[map[string]T]),
	}

	// Create a default configuration, and then apply the options.
	cfg := &Config{
		clock:            NewClock(),
		evictionInterval: ttl / time.Duration(numShards),
		getSize:          client.Size,
		log:              slog.Default(),
	}
	// Apply the options to the configuration.
	client.Config = cfg
	for _, opt := range opts {
		opt(cfg)
	}
	validateConfig(capacity, numShards, ttl, evictionPercentage, cfg)

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

// Get retrieves a single value from the cache.
func (c *Client[T]) Get(key string) (T, bool) {
	shard := c.getShard(key)
	val, ok, ignore, _ := shard.get(key)
	c.reportCacheHits(ok)
	return val, ok && !ignore
}

// GetMany retrieves multiple values from the cache.
func (c *Client[T]) GetMany(keys []string) map[string]T {
	records := make(map[string]T, len(keys))
	for _, key := range keys {
		if value, ok := c.Get(key); ok {
			records[key] = value
		}
	}
	return records
}

// GetManyKeyFn follows the same API as GetFetchBatch and PassthroughBatch.
// You provide it with a slice of IDs and a keyFn, which is applied to create
// the cache key. The returned map uses the IDs as keys instead of the cache key.
// If you've used ScanKeys to retrieve the actual keys, you can retrieve the records
// using GetMany instead.
func (c *Client[T]) GetManyKeyFn(ids []string, keyFn KeyFn) map[string]T {
	records := make(map[string]T, len(ids))
	for _, id := range ids {
		if value, ok := c.Get(keyFn(id)); ok {
			records[id] = value
		}
	}
	return records
}

// Set writes a single value to the cache. Returns true if it triggered an eviction.
func (c *Client[T]) Set(key string, value T) bool {
	shard := c.getShard(key)
	return shard.set(key, value, false)
}

// StoreMissingRecord writes a single value to the cache. Returns true if it triggered an eviction.
func (c *Client[T]) StoreMissingRecord(key string) bool {
	shard := c.getShard(key)
	var zero T
	return shard.set(key, zero, true)
}

// SetMany writes a map of key value pairs to the cache.
func (c *Client[T]) SetMany(records map[string]T) bool {
	var triggeredEviction bool
	for key, value := range records {
		evicted := c.Set(key, value)
		if evicted {
			triggeredEviction = true
		}
	}
	return triggeredEviction
}

// SetManyKeyFn follows the same API as GetFetchBatch and PassThroughBatch. It
// takes a map of records where the keyFn is applied to each key in the map
// before it's stored in the cache.
func (c *Client[T]) SetManyKeyFn(records map[string]T, cacheKeyFn KeyFn) bool {
	var triggeredEviction bool
	for id, value := range records {
		evicted := c.Set(cacheKeyFn(id), value)
		if evicted {
			triggeredEviction = true
		}
	}
	return triggeredEviction
}

// ScanKeys returns a list of all keys in the cache.
func (c *Client[T]) ScanKeys() []string {
	keys := make([]string, 0, c.Size())
	for _, shard := range c.shards {
		keys = append(keys, shard.keys()...)
	}
	return keys
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

// NumKeysInflight returns the number of keys that are currently being fetched.
func (c *Client[T]) NumKeysInflight() int {
	c.inFlightMutex.Lock()
	defer c.inFlightMutex.Unlock()
	c.inFlightBatchMutex.Lock()
	defer c.inFlightBatchMutex.Unlock()
	return len(c.inFlightMap) + len(c.inFlightBatchMap)
}
