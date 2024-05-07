# `sturdyc`: a caching library for building sturdy systems

[![Go Reference](https://pkg.go.dev/badge/github.com/creativecreature/sturdyc.svg)](https://pkg.go.dev/github.com/creativecreature/sturdyc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/sturdyc/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativecreature/sturdyc)](https://goreportcard.com/report/github.com/creativecreature/sturdyc)
[![Test](https://github.com/creativecreature/sturdyc/actions/workflows/main.yml/badge.svg)](https://github.com/creativecreature/sturdyc/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/creativecreature/sturdyc/graph/badge.svg?token=CYSKW3Z7E6)](https://codecov.io/gh/creativecreature/sturdyc)

`Sturdyc` is a highly concurrent cache that supports non-blocking reads and has
a configurable number of shards that makes it possible to achieve parallel
writes. The [xxhash](https://github.com/cespare/xxhash) algorithm is used for
efficient key distribution. Evictions are performed per shard based on recency at
O(N) time complexity using [quickselect](https://en.wikipedia.org/wiki/Quickselect).

It has all the functionality you would expect from a caching library, along
with additional features designed to help you build performant and robust
applications, such as:

- [**stampede protection**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#stampede-protection)
- [**caching non-existent records**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#non-existent-records)
- [**caching batch endpoints per record**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#batch-endpoints)
- [**cache key permutations**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#cache-key-permutations)
- [**refresh buffering**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#refresh-buffering)
- [**request passthrough**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#passthrough)
- [**custom metrics**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#custom-metrics)
- [**generics**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#generics)

The package has great support for batchable data sources as it takes the
response apart, and then caches each record individually based on the
permutations of the options with which it was fetched. It can also
significantly reduce the traffic to the underlying data sources by creating a
buffer for each unique option set, and then delaying the refreshes until it has
gathered enough IDs

Records can be configured to refresh either based on time or at a certain rate
of requests. All refreshes occur in the background, ensuring that users never
have to wait for a record to be updated, resulting in _very low latency_
applications while also allowing unused keys to expire.

There are examples for all of these configurations further down this file!

# Installing

```sh
go get github.com/creativecreature/sturdyc
```

# At a glance

The package exports the following functions:

- Use [`Get`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Get) to get a record from the cache.
- Use [`GetFetch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.GetFetch) to have the cache fetch and store a record.
- Use [`GetFetchBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.GetFetchBatch) to have the cache fetch and store a batch of records.
- Use [`Set`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Set) to write a record to the cache.
- Use [`SetMany`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.SetMany) to write multiple records to the cache.
- Use [`Delete`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Delete) to delete a record from the cache.
- Use [`Passthrough`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Passthrough) to have the cache fetch and store a record.
- Use [`PassthroughBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.PassthroughBatch) to have the cache fetch and store a batch of records.
- Use [`Size`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Size) to get the number of items in the cache.
- Use [`Size`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.Size) to get the number of items in the cache.
- Use [`PermutatedKey`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.PermutatedKey) to create a permutated cache key.
- Use [`PermutatedBatchKeyFn`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.PermutatedBatchKey) to create a permutated cache key for every record in a batch.
- Use [`BatchKeyFn`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Client.BatchKeyFn) to create a cache key for every record in a batch.

To utilize these functions, you will first have to set up a cache client to
manage your configuration:

```go
	// Maximum number of entries in the cache.
	capacity := 10000
	// Number of shards to use for the sturdyc.
	numShards := 10
	// Time-to-live for cache entries.
	ttl := 2 * time.Hour
	// Percentage of entries to evict when the cache is full. Setting this
	// to 0 will make set a no-op if the cache has reached its capacity.
	evictionPercentage := 10

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[int](capacity, numShards, ttl, evictionPercentage)

	cacheClient.Set("key1", 99)
	log.Println(cacheClient.Size())
	log.Println(cacheClient.Get("key1"))

	cacheClient.Delete("key1")
	log.Println(cacheClient.Size())
	log.Println(cacheClient.Get("key1"))
```

Next, we'll look at some of the more _advanced features_.

# Stampede protection

Cache stampedes (also known as thundering herd) occur when many requests for a
particular piece of data (which has just expired or been evicted from the
cache) come in at once. This can cause all requests to fetch the data
concurrently, subsequently causing a significant load on the underlying data
source.

To prevent this, we can enable **stampede protection**:

```go
func main() {
	// Set a minimum and maximum refresh delay for the sturdyc. This is
	// used to spread out the refreshes for our entries evenly over time.
	minRefreshDelay := time.Millisecond * 10
	maxRefreshDelay := time.Millisecond * 30
	// The base used for exponential backoff when retrying a refresh.
	retryBaseDelay := time.Millisecond * 10
	// NOTE: Ignore this for now, it will be shown in the next example.
	storeMisses := false

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
	)
}
```

With this configuration, the cache will prevent records from expiring by
enqueueing refreshes when a key is requested again at a random interval between
10 and 30 milliseconds. Performing the refreshes upon cache key retrieval,
rather than at a fixed interval, allows unused keys to expire.

To demonstrate this, we can create a simple API client:

```go
type API struct {
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c}
}

func (a *API) Get(ctx context.Context, key string) (string, error) {
	// This could be a call to a rate limited service, a database query, etc.
	fetchFn := func(_ context.Context) (string, error) {
		log.Printf("Fetching value for key: %s\n", key)
		return "value", nil
	}
	return a.GetFetch(ctx, key, fetchFn)
}
```

and use it in our `main` function:

```go
func main() {
	// ...

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
	)

	// Create a new API instance with the cache client.
	api := NewAPI(cacheClient)

	// We are going to retrieve the values every 10 milliseconds, however the
	// logs will reveal that actual refreshes fluctuate randomly within a 10-30
	// millisecond range. Even if this loop is executed across multiple
	// goroutines, the API call frequency will maintain this variability,
	// ensuring we avoid overloading the API with requests.
	for i := 0; i < 100; i++ {
		val, _ := api.Get(context.Background(), "key")
		log.Printf("Value: %s\n", val)
		time.Sleep(minRefreshDelay)
	}
}
```

Running this program, we're going to see that the value gets refreshed once
every 2-3 retrievals:

```sh
go run .
2024/04/07 09:05:29 Fetching value for key: key
2024/04/07 09:05:29 Value: value
2024/04/07 09:05:29 Value: value
2024/04/07 09:05:29 Fetching value for key: key
2024/04/07 09:05:29 Value: value
2024/04/07 09:05:29 Value: value
2024/04/07 09:05:29 Value: value
2024/04/07 09:05:29 Fetching value for key: key
...
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/stampede)

# Non-existent records

Another factor to consider is non-existent keys. It could be an ID that has
been added manually to a CMS with a typo that leads to no data being returned
from the upstream source. This can significantly increase our systems latency,
as we're never able to get a cache hit and serve from memory.

However, it could also be caused by a slow ingestion process. Perhaps it takes
some time for a new entity to propagate through a distributed system.

The cache allows us to store these IDs as missing records, while refreshing
them like any other entity. To illustrate, we can extend the previous example
to enable this functionality:

```go
func main() {
	// ...

	// Tell the cache to store missing records.
	storeMisses := true

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
	)

	// Create a new API instance with the cache client.
	api := NewAPI(cacheClient)

	// ...
	for i := 0; i < 100; i++ {
		val, err := api.Get(context.Background(), "key")
		if errors.Is(err, sturdyc.ErrMissingRecord) {
			log.Println("Record does not exist.")
		}
		if err == nil {
			log.Printf("Value: %s\n", val)
		}
		time.Sleep(minRefreshDelay)
	}
}
```

Next, we'll modify the API client to return the `ErrStoreMissingRecord` error
for the first _3_ calls. This error instructs the cache to store it as a missing
record:

```go
type API struct {
	*sturdyc.Client[string]
	count int
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c, 0}
}

func (a *API) Get(ctx context.Context, key string) (string, error) {
	fetchFn := func(_ context.Context) (string, error) {
		log.Printf("Fetching value for key: %s\n", key)
		a.count++
		if a.count < 3 {
			return "", sturdyc.ErrStoreMissingRecord
		}
		return "value", nil
	}
	return a.GetFetch(ctx, key, fetchFn)
}
```

and then call it:

```go
func main() {
	// ...

	for i := 0; i < 100; i++ {
		val, err := api.Get(context.Background(), "key")
		if errors.Is(err, sturdyc.ErrMissingRecord) {
			log.Println("Record does not exist.")
		}
		if err == nil {
			log.Printf("Value: %s\n", val)
		}
		time.Sleep(minRefreshDelay)
	}
}
```

Running this program, we'll see that the record is missing during the first 3
refreshes and then transitions into having a value:

```sh
2024/04/07 09:42:49 Fetching value for key: key
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Fetching value for key: key
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Record does not exist.
2024/04/07 09:42:49 Fetching value for key: key
2024/04/07 09:42:49 Value: value
2024/04/07 09:42:49 Value: value
2024/04/07 09:42:49 Fetching value for key: key
...
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/missing)

# Batch endpoints

One challenge with caching batchable endpoints is that you have to find a way
to reduce the number of keys. To illustrate, let's say that we have 10 000
records, and an endpoint for fetching them that allows for batches of 20.

Now, let's calculate the number of combinations if we were to create the cache
keys from the query params:

$$ C(n, k) = \binom{n}{k} = \frac{n!}{k!(n-k)!} $$

For $n = 10,000$ and $k = 20$, this becomes:

$$ C(10,000, 20) = \binom{10,000}{20} = \frac{10,000!}{20!(10,000-20)!} $$

This results in an approximate value of:

$$ \approx 4.032 \times 10^{61} $$

and this is if we're sending perfect batches of 20. If we were to do 1 to 20
IDs (not just exactly 20 each time) the total number of combinations would be
the sum of combinations for each k from 1 to 20.

`sturdyc` caches each record individually, which effectively prevents factorial
increases in cache keys.

To see how this works, we can look at a small example application. This time,
we'll start with the API client:

```go
type API struct {
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c}
}

func (a *API) GetBatch(ctx context.Context, ids []string) (map[string]string, error) {
	// We are going to pass the cache a key function that prefixes each id.
	// This makes it possible to save the same id for different data sources.
	cacheKeyFn := a.BatchKeyFn("some-prefix")

	// The fetchFn is only going to retrieve the IDs that are not in the cache.
	fetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		log.Printf("Cache miss. Fetching ids: %s\n", strings.Join(cacheMisses, ", "))
		// Batch functions should return a map where the key is the id of the record.
		// If you have storage of missing records enabled, any ID that isn't present
		// in this map is going to be stored as a cache miss.
		response := make(map[string]string, len(cacheMisses))
		for _, id := range cacheMisses {
			response[id] = "value"
		}
		return response, nil
	}

	return a.GetFetchBatch(ctx, ids, cacheKeyFn, fetchFn)
}
```

and we're going to use the same cache configuration as the previous example, so
I've omitted it for brevity:

```go
func main() {
	// ...

	// Create a new API instance with the cache client.
	api := NewAPI(cacheClient)

	// Seed the cache with ids 1-10.
	log.Println("Seeding ids 1-10")
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	api.GetBatch(context.Background(), ids)
	log.Println("Seed completed")

	// Each record has been cached individually. To illustrate this, we can keep
	// fetching a random number of records from the original batch, plus a new ID.
	// Looking at the logs, we'll should be able to see that the cache only
	// fetches the id that wasn't present in the original batch.
	for i := 1; i <= 100; i++ {
		// Get N ids from the original batch.
		recordsToFetch := rand.IntN(10) + 1
		batch := make([]string, recordsToFetch)
		copy(batch, ids[:recordsToFetch])
		// Add a random ID between 1 and 100 to the batch.
		batch = append(batch, strconv.Itoa(rand.IntN(1000)+10))
		values, _ := api.GetBatch(context.Background(), batch)
		// Print the records we retrieved from the cache.
		log.Println(values)
	}
}
```

Running this code, we can see that we only end up fetching the randomized ID,
while continuously getting cache hits for IDs 1-10, regardless of what the
batch looks like:

```sh
...
2024/04/07 11:09:58 Seed completed
2024/04/07 11:09:58 Cache miss. Fetching ids: 173
2024/04/07 11:09:58 map[1:value 173:value 2:value 3:value 4:value]
2024/04/07 11:09:58 Cache miss. Fetching ids: 12
2024/04/07 11:09:58 map[1:value 12:value 2:value 3:value 4:value]
2024/04/07 11:09:58 Cache miss. Fetching ids: 730
2024/04/07 11:09:58 map[1:value 2:value 3:value 4:value 730:value]
2024/04/07 11:09:58 Cache miss. Fetching ids: 520
2024/04/07 11:09:58 map[1:value 2:value 3:value 4:value 5:value 520:value 6:value 7:value 8:value]
...
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/batch)

# Cache key permutations

If you're attempting to cache data from an upstream system, the ID alone may be
insufficient to uniquely identify the record in your cache. The endpoint you're
calling might accept a variety of options that transform the data in different
ways. Therefore, it's important to consider this and store records for each
unique set of options.

To showcase this, we can create a simple API client that interacts with a
service for retrieving order statuses:

```go
type OrderOptions struct {
	CarrierName        string
	LatestDeliveryTime string
}

type OrderAPI struct {
	*sturdyc.Client[string]
}

func NewOrderAPI(c *sturdyc.Client[string]) *OrderAPI {
	return &OrderAPI{c}
}

func (a *OrderAPI) OrderStatus(ctx context.Context, ids []string, opts OrderOptions) (map[string]string, error) {
	// We use the  PermutedBatchKeyFn when an ID isn't enough to uniquely identify a
	// record. The cache is going to store each id once per set of options. In a more
	// realistic scenario, the opts would be query params or arguments to a DB query.
	cacheKeyFn := a.PermutatedBatchKeyFn("key", opts)

	// We'll create a fetchFn with a closure that captures the options. For this
	// simple example, it logs and returns the status for each order, but you could
	// just as easily have called an external API.
	fetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		log.Printf("Fetching: %v, carrier: %s, delivery time: %s\n", cacheMisses, opts.CarrierName, opts.LatestDeliveryTime)
		response := map[string]string{}
		for _, id := range cacheMisses {
			response[id] = fmt.Sprintf("Available for %s", opts.CarrierName)
		}
		return response, nil
	}
	return a.GetFetchBatch(ctx, ids, cacheKeyFn, fetchFn)
}
```

The main difference from the previous example is that we're using the
`PermutatedBatchKeyFn` function. Internally, the cache uses reflection to
extract the names and values of every exported field, and uses them to build
the cache keys.

To demonstrate this, we can write another small program:

```go
func main() {
	// ...

	// Create a new cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
	)

	// We will fetch these IDs using various option sets.
	ids := []string{"id1", "id2", "id3"}
	optionSetOne := OrderOptions{CarrierName: "FEDEX", LatestDeliveryTime: "2024-04-06"}
	optionSetTwo := OrderOptions{CarrierName: "DHL", LatestDeliveryTime: "2024-04-07"}
	optionSetThree := OrderOptions{CarrierName: "UPS", LatestDeliveryTime: "2024-04-08"}

	orderClient := NewOrderAPI(cacheClient)
	ctx := context.Background()

	// Next, we'll seed our cache by fetching the entire list of IDs for all options sets.
	log.Println("Filling the cache with all IDs for all option sets")
	orderClient.OrderStatus(ctx, ids, optionSetOne)
	orderClient.OrderStatus(ctx, ids, optionSetTwo)
	orderClient.OrderStatus(ctx, ids, optionSetThree)
	log.Println("Cache filled")
}
```

At this point, the cache has stored each record individually for each option set:

- FEDEX-2024-04-06-id1
- DHL-2024-04-07-id1
- UPS-2024-04-08-id1
- etc..

Next, we'll add a sleep to make sure that all of the records are due for a
refresh, and then request the ids individually for each set of options:

```go
func main() {
	// ...

	// Sleep to make sure that all records are due for a refresh.
	time.Sleep(maxRefreshDelay + 1)

	// Fetch each id for each option set.
	for i := 0; i < len(ids); i++ {
		orderClient.OrderStatus(ctx, []string{ids[i]}, optionSetOne)
		orderClient.OrderStatus(ctx, []string{ids[i]}, optionSetTwo)
		orderClient.OrderStatus(ctx, []string{ids[i]}, optionSetThree)
	}

	// Sleep for a second to allow the refresh logs to print.
	time.Sleep(time.Second)
}
```

Running this program, we can see that the records are refreshed once per unique
id+option combination:

```sh
go run .
2024/04/07 13:33:56 Filling the cache with all IDs for all option sets
2024/04/07 13:33:56 Fetching: [id1 id2 id3], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:56 Fetching: [id1 id2 id3], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:33:56 Fetching: [id1 id2 id3], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:56 Cache filled
2024/04/07 13:33:58 Fetching: [id3], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id1], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id1], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id2], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id2], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id3], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id2], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:33:58 Fetching: [id3], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:33:58 Fetching: [id1], carrier: DHL, delivery time: 2024-04-07
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/permutations)

# Refresh buffering

As seen in the example above, we're not really utilizing the fact that the
endpoint is batchable when we're performing the refreshes.

To make this more efficient, we can enable the _refresh buffering_
functionality. Internally, the cache is going to create a buffer for each
permutation of our options. It is then going to collect ids until it reaches a
certain size, or exceeds a time threshold.

The only change we have to make to the example above is to enable this
functionality:

```go
func main() {
	// ...

	// With refresh buffering enabled, the cache will buffer refreshes
	// until the batch size is reached or the buffer timeout is hit.
	batchSize := 3
	batchBufferTimeout := time.Second * 30

	// Create a new cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
		sturdyc.WithRefreshBuffering(batchSize, batchBufferTimeout),
	)

	// ...
}
```

and now we can see that the cache performs the refreshes in batches per
permutation of our query params:

```sh
go run .
2024/04/07 13:45:42 Filling the cache with all IDs for all option sets
2024/04/07 13:45:42 Fetching: [id1 id2 id3], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:45:42 Fetching: [id1 id2 id3], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:45:42 Fetching: [id1 id2 id3], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:45:42 Cache filled
2024/04/07 13:45:44 Fetching: [id1 id2 id3], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:45:44 Fetching: [id1 id3 id2], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:45:44 Fetching: [id1 id2 id3], carrier: UPS, delivery time: 2024-04-08
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/buffering)

# Passthrough

Time-based refreshes work really well for most use cases. However, there are
scenarios where you might want to allow a certain amount of traffic to hit the
underlying data source. For example, you might achieve a 99.99% cache hit rate,
and even though you refresh the data every 1-2 seconds, it only amounts to a
handful of requests. This could cause the other system to scale down too much.

To solve this problem, `sturdyc` provides you with `client.Passthrough` and
`client.PassthroughBatch`. These functions are functionally equivalent to
`client.GetFetch` and `client.GetFetchBatch`, except they allow a certain
percentage of requests through rather than allowing a request through every x
milliseconds/seconds/minutes/hours.

The passthroughs will still be performed in the background, which means that
your application will maintain low latency response times. Moreover, if the
underlying system goes down, `client.Passthrough` and `client.PassthroughBatch`
will still be able to serve stale data until the record's TTL expires.

```go
capacity := 5
numShards := 2
ttl := time.Minute
evictionPercentage := 10
c := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
    // Allow 50% of the requests to pass-through. Default is 100%.
    sturdyc.WithPassthroughPercentage(50),
    // Buffer the batchable pass-throughs. Default is false.
    sturdyc.WithPassthroughBuffering(),
)

res, err := sturdyc.Passthrough(ctx, c, "id", fetchFn)
batchRes, batchErr := sturdyc.PassthroughBatch(ctx, c, idBatch, c.BatchKeyFn("item"), batchFetchFn)
```

# Custom metrics

The cache can be configured to report custom metrics for:

- Size of the cache
- Cache hits
- Cache misses
- Evictions
- Forced evictions
- The number of entries evicted
- Shard distribution
- The size of the refresh buckets

All you have to do is implement the `MetricsRecorder` interface:

```go
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
```

and pass it as an option when you create the client:

```go
cache := sturdyc.New[any](
    cacheSize,
    shardSize,
    cacheTTL,
    evictWhenFullPercentage,
    sturdyc.WithMetrics(metricsRecorder),
)
```

Below are a few images where these metrics have been visualized in Grafana:

<img width="939" alt="Screenshot 2024-05-04 at 12 36 43" src="https://github.com/creativecreature/sturdyc/assets/12787673/1f630aed-2322-4d3a-9510-d582e0294488">

<img width="942" alt="Screenshot 2024-05-04 at 12 37 39" src="https://github.com/creativecreature/sturdyc/assets/12787673/25187529-28fb-4c4e-8fe9-9fb48772e0c0">

<img width="941" alt="Screenshot 2024-05-04 at 12 38 04" src="https://github.com/creativecreature/sturdyc/assets/12787673/b1359867-f1ef-4a09-8c75-d7d2360726f1">

<img width="940" alt="Screenshot 2024-05-04 at 12 38 20" src="https://github.com/creativecreature/sturdyc/assets/12787673/de7f00ee-b14d-443b-b69e-91e19665c252">

# Generics

There are scenarios, where you might want to use the same cache for a data
source that could return multiples types. Personally, I tend to create caches
based on how frequently the data needs to be refreshed. I'll often have one
transient cache which refreshes the data every 2-5 milliseconds. Hence, I'll
use `any` as the type:

```go
	cacheClient := sturdyc.New[any](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithStampedeProtection(minRefreshDelay, maxRefreshDelay, retryBaseDelay, storeMisses),
		sturdyc.WithRefreshBuffering(10, time.Second*15),
	)
```

However, if this data source has more than a handful of types, the type
conversions may quickly feel like too much boilerplate. If that is the case,
you can use any of these package level functions:

- [`GetFetch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#GetFetch)
- [`GetFetchBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#GetFetchBatch)
- [`Passthrough`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Passthrough)
- [`PassthroughBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#PassthroughBatch)

They will perform the type conversions for you, and if any of them where to fail,
you'll get a [`ErrInvalidType`](https://pkg.go.dev/github.com/creativecreature/sturdyc#pkg-variables) error.

Below is an example of what an API client that uses these functions might look
like:

```go
type OrderAPI struct {
	cacheClient *sturdyc.Client[any]
}

func NewOrderAPI(c *sturdyc.Client[any]) *OrderAPI {
	return &OrderAPI{cacheClient: c}
}

func (a *OrderAPI) OrderStatus(ctx context.Context, ids []string) (map[string]string, error) {
	cacheKeyFn := a.cacheClient.BatchKeyFn("order-status")
	fetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		response := make(map[string]string, len(ids))
		for _, id := range cacheMisses {
			response[id] = "Order status: pending"
		}
		return response, nil
	}
	return sturdyc.GetFetchBatch(ctx, a.cacheClient, ids, cacheKeyFn, fetchFn)
}

func (a *OrderAPI) DeliveryTime(ctx context.Context, ids []string) (map[string]time.Time, error) {
	cacheKeyFn := a.cacheClient.BatchKeyFn("delivery-time")
	fetchFn := func(_ context.Context, cacheMisses []string) (map[string]time.Time, error) {
		response := make(map[string]time.Time, len(ids))
		for _, id := range cacheMisses {
			response[id] = time.Now()
		}
		return response, nil
	}
	return sturdyc.GetFetchBatch(ctx, a.cacheClient, ids, cacheKeyFn, fetchFn)
}
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/generics)
