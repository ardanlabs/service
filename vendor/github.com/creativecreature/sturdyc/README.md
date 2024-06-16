# `sturdyc`: a caching library for building sturdy systems

[![Go Reference](https://pkg.go.dev/badge/github.com/creativecreature/sturdyc.svg)](https://pkg.go.dev/github.com/creativecreature/sturdyc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/sturdyc/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativecreature/sturdyc)](https://goreportcard.com/report/github.com/creativecreature/sturdyc)
[![Test](https://github.com/creativecreature/sturdyc/actions/workflows/main.yml/badge.svg)](https://github.com/creativecreature/sturdyc/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/creativecreature/sturdyc/graph/badge.svg?token=CYSKW3Z7E6)](https://codecov.io/gh/creativecreature/sturdyc)

`Sturdyc` is a highly concurrent cache that supports **non-blocking reads** and has
a configurable number of shards that makes it possible to achieve writes
**without any lock contention**.

The [xxhash](https://github.com/cespare/xxhash) algorithm is used for efficient
key distribution.

The cache performs continuous evictions of each shard. There are options to
both disable this functionality and tweak the interval. When you create a
cache client, you get to decide the percentage of records to evict if the
capacity is reached.

All evictions are performed per shard based on recency, with an _O(N) time
complexity_, using [quickselect](https://en.wikipedia.org/wiki/Quickselect).

It has all the functionality you would expect from a caching library, but what
**sets it apart** are the features designed to make I/O heavy applications both
_robust_ and _highly performant_.

### Adding `sturdyc` to your application:

To illustrate how to integrate this package with your application, we'll use
the following two methods of an API client as example:

```go
// Order retrieves a single order by ID.
func (c *Client) Order(ctx context.Context, id string) (Order, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var response Order
	err := requests.URL(c.orderURL).
		Pathf("/order/%s", id).
		ToJSON(&response).
		Fetch(timeoutCtx)

	return response, err
}

// Orders retrieves a batch of orders by their IDs.
func (c *Client) Orders(ctx context.Context, ids []string) (map[string]Order, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var response map[string]Order
	err := requests.URL(c.orderURL).
		Path("/orders").
		Param("ids", strings.Join(ids, ",")).
		ToJSON(&response).
		Fetch(timeoutCtx)

	return response, err
}
```

Now, all we have to do is wrap the fetching part in a function and then hand it
over to the cache:

```go
func (c *Client) Order(ctx context.Context, id string) (Order, error) {
	fetchFunc := func(ctx context.Context) (Order, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		var response Order
		err := requests.URL(c.orderURL).
			Pathf("/order/%s", id).
			ToJSON(&response).
			Fetch(timeoutCtx)

		return response, err
	}

	return c.cache.GetOrFetch(ctx, "order-"+id, fetchFunc)
}

func (c *Client) Orders(ctx context.Context, ids []string) (map[string]Order, error) {
	fetchFunc := func(ctx context.Context, cacheMisses []string) (map[string]Order, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		var response map[string]Order
		err := requests.URL(c.orderURL).
			Path("/orders").
			Param("ids", strings.Join(cacheMisses, ",")).
			ToJSON(&response).
			Fetch(timeoutCtx)

		return response, err
	}

	return c.cache.GetOrFetchBatch(ctx, ids, c.persistentCache.BatchKeyFn("orders"), fetchFunc)
}
```

The example above retrieves the data from an HTTP API, but it's just as easy to
wrap a database query, a remote procedure call, a disk read, or any other I/O
operation

These three extra lines of code will obviously grant us the ability to serve
the data from memory, and then retrieve it again once the TTL expires, but the
cache can do much more. Let's look at that next!

### Benefits:

#### Deduplication

When we pass our functions for data retrieval to `sturdyc`, it will
automatically perform _in-flight_ tracking for every key. This also works for
batch operations, where it can deduplicate a batch of cache misses and then
assemble the response by picking records from multiple in-flight requests.

#### Early refreshes

There is also a lot of extra functionality you can enable, one being _early
refreshes_ which instructs the cache to refresh the keys which are in active
rotation, thereby preventing them from ever expiring. This can have a huge
impact on an applications latency as you're able to continiously serve the most
frequently used data from memory:

```go
sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, exponentialBackOff)
```

#### Batching

When the cache retrieves data from a batchable source, it will disassemble the
response and then cache each record individually based on the permutations of
the options with which it was fetched.

We can leverage this fact to **significantly reduce** our application's
outgoing requests to these data sources by enabling _refresh coalescing_.
Internally, `sturdyc` creates a buffer for each option set and gathers IDs
until the `idealBatchSize` is reached or the `batchBufferTimeout` expires:

```go
sturdyc.WithRefreshCoalescing(idealBatchSize, batchBufferTimeout)
```

#### Distributed key-value store

You can also configure `sturdyc` to synchronize its in-memory cache with a
**distributed key-value store** of your choosing:

```go
sturdyc.WithDistributedStorage(storage),
```

#### Latency improvements

Below is a screenshot showing the latency improvements we've observed after
replacing our old cache with this package:

&nbsp;
<img width="1554" alt="Screenshot 2024-05-10 at 10 15 18" src="https://github.com/creativecreature/sturdyc/assets/12787673/adad1d4c-e966-4db1-969a-eda4fd75653a">
&nbsp;

In addition to this, we've seen our number of outgoing requests decrease by
more than 90% while still serving data that is refreshed every second. This
setting is configurable, and you can adjust it to a lower value if you like.

### Table of contents

There are examples further down this file that covers the entire API, and I
encourage you to **read these examples in the order they appear**. Most of them
build on each other, and many share configurations. Here is a brief overview of
what the examples are going to cover:

- [**stampede protection**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#stampede-protection)
- [**early refreshes**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#early-refreshes)
- [**caching non-existent records**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#non-existent-records)
- [**caching batch endpoints per record**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#batch-endpoints)
- [**cache key permutations**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#cache-key-permutations)
- [**refresh coalescing**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#refresh-coalescing)
- [**request passthrough**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#passthrough)
- [**distributed storage**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#distributed-storage)
- [**custom metrics**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#custom-metrics)
- [**generics**](https://github.com/creativecreature/sturdyc?tab=readme-ov-file#generics)

# Installing

```sh
go get github.com/creativecreature/sturdyc
```

# Getting started

The first thing you will have to do is to create a cache client to hold your
configuration:

```go
	// Maximum number of entries in the cache. Exceeding this number will trigger
	// an eviction (as long as the "evictionPercentage" is greater than 0).
	capacity := 10000
	// Number of shards to use. Increasing this number will reduce write lock collisions.
	numShards := 10
	// Time-to-live for cache entries.
	ttl := 2 * time.Hour
	// Percentage of entries to evict when the cache reaches its capacity. Setting this
	// to 0 will make writes a no-op until an item has either expired or been deleted.
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

Next, we'll look at some of the more _advanced features_ in detail.

# Stampede protection

Cache stampedes (also known as thundering herd) occur when many requests for a
particular piece of data, which has just expired or been evicted from the
cache, come in at once

Preventing this has been one of the key objectives for this package. We do not
want to cause a significant load on the underlying data source every time a key
expires.

The `GetOrFetch` function takes a key and a function for retrieving the data if
it's not in the cache. The cache is going to ensure that we never have more
than a single request per key. It achieves this by tracking all of the
in-flight requests:

```go
	var count atomic.Int32
	fetchFn := func(_ context.Context) (int, error) {
		count.Add(1)
		time.Sleep(time.Second)
		return 1337, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			// We can ignore the error given the fetchFn we're using.
			val, _ := cacheClient.GetOrFetch(context.Background(), "key2", fetchFn)
			log.Printf("got value: %d\n", val)
			wg.Done()
		}()
	}
	wg.Wait()

	log.Printf("fetchFn was called %d time\n", count.Load())
	log.Println(cacheClient.Get("key2"))

```

Running this program we'll see that our requests for "key2" were deduplicated,
and that the fetchFn only got called once:

```sh
❯ go run .
2024/05/21 08:06:29 got value: 1337
2024/05/21 08:06:29 got value: 1337
2024/05/21 08:06:29 got value: 1337
2024/05/21 08:06:29 got value: 1337
2024/05/21 08:06:29 got value: 1337
2024/05/21 08:06:29 fetchFn was called 1 time
2024/05/21 08:06:29 1337 true
```

We can use the `GetOrFetchBatch` function for data sources that supports batching.
To demonstrate this, I'll create a mock function that sleeps for `5` seconds,
and then returns a map with a numerical value for every ID:

```go
	var count atomic.Int32
	fetchFn := func(_ context.Context, ids []string) (map[string]int, error) {
		count.Add(1)
		time.Sleep(time.Second * 5)

		response := make(map[string]int, len(ids))
		for _, id := range ids {
			num, _ := strconv.Atoi(id)
			response[id] = num
		}

		return response, nil
	}
```

Next, we'll need some batches to test with, so I created three batches with 5
IDs each:

```go
	batches := [][]string{
		{"1", "2", "3", "4", "5"},
		{"6", "7", "8", "9", "10"},
		{"11", "12", "13", "14", "15"},
	}
```

IDs can often be fetched from multiple data sources. Hence, we'll want to
prefix the ID in order to make the cache key unique. The package provides more
functionality for this that we'll see later on, but for now we'll use the most
simple version which adds a string prefix to every ID:

```go
	keyPrefixFn := cacheClient.BatchKeyFn("my-data-source")
```

we can now request each batch in a separate goroutine:

```go
	for _, batch := range batches {
		go func() {
			res, _ := cacheClient.GetOrFetchBatch(context.Background(), batch, keyPrefixFn, fetchFn)
			log.Printf("got batch: %v\n", res)
		}()
	}

	// Give the goroutines above a chance to run to ensure that the batches are in-flight.
	time.Sleep(time.Second * 3)
```

At this point, the cache should have in-flight requests for IDs 1-15. Knowing
this, we'll test the stampede protection by launching another five goroutines.
Each goroutine is going to request two random IDs from our batches:

```go
	// Launch another 5 goroutines that are going to pick two random IDs from any of the batches.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			ids := []string{batches[rand.IntN(2)][rand.IntN(4)], batches[rand.IntN(2)][rand.IntN(4)]}
			res, _ := cacheClient.GetOrFetchBatch(context.Background(), ids, keyPrefixFn, fetchFn)
			log.Printf("got batch: %v\n", res)
			wg.Done()
		}()
	}

	wg.Wait()
	log.Printf("fetchFn was called %d times\n", count.Load())
```

Running this program, and looking at the logs, we'll see that the cache is able
to pick IDs from different batches:

```sh
❯ go run .
2024/05/21 09:14:23 got batch: map[8:8 9:9]
2024/05/21 09:14:23 got batch: map[4:4 9:9] <---- NOTE: ID 4 and 9 are part of different batches
2024/05/21 09:14:23 got batch: map[11:11 12:12 13:13 14:14 15:15]
2024/05/21 09:14:23 got batch: map[1:1 7:7] <---- NOTE: ID 1 and 7 are part of different batches
2024/05/21 09:14:23 got batch: map[10:10 6:6 7:7 8:8 9:9]
2024/05/21 09:14:23 got batch: map[3:3 9:9] <---- NOTE: ID 3 and 9 are part of different batches
2024/05/21 09:14:23 got batch: map[1:1 2:2 3:3 4:4 5:5]
2024/05/21 09:14:23 got batch: map[4:4 9:9] <---- NOTE: ID 4 and 9 are part of different batches
2024/05/21 09:14:23 fetchFn was called 3 times <---- NOTE: We only generated 3 outgoing requests.
```

And on the last line, we can see that the additional calls didn't generate any
further outgoing requests. The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/basic)

## Early refreshes

Being able to prevent your most frequently used records from ever expiring can
have a significant impact on your application's latency. Therefore, the package
provides a `WithEarlyRefreshes` option, which instructs the cache to
continuously refresh these records in the background.

A refresh gets scheduled if a key is **requested again** after a configurable
amount of time has passed. This is an important distinction because it means
that the cache doesn't just naively refresh every key it's ever seen. Instead,
it only refreshes the records that are actually in rotation, while allowing
unused keys to be deleted once their TTL expires.

Below is an example configuration:

```go
func main() {
	// Set a minimum and maximum refresh delay for the record. This is
	// used to spread out the refreshes of our entries evenly over time.
	// We don't want our outgoing requests graph to look like a comb that
    // sends a spike of refreshes every 30 ms.
	minRefreshDelay := time.Millisecond * 10
	maxRefreshDelay := time.Millisecond * 30
	// The base used for exponential backoff when retrying a refresh. Most of the
	// time, we perform refreshes well in advance of the records expiry time.
	// Hence, we can use this to make it easier for a system that is having
	// trouble to get back on it's feet by making fewer refreshes when we're
	// seeing a lot of errors. Once we receive a successful response, the
	// refreshes return to their original frequency. You can set this to 0
	// if you don't want this behavior.
	retryBaseDelay := time.Millisecond * 10

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
	)
}
```

To get a feeling for how this works, we'll create a simple API client that
embedds the cache:

```go
type API struct {
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c}
}

func (a *API) Get(ctx context.Context, key string) (string, error) {
	// This could be an API call, a database query, etc.
    fetchFn := func(_ context.Context) (string, error) {
		log.Printf("Fetching value for key: %s\n", key)
		return "value", nil
	}
	return a.GetOrFetch(ctx, key, fetchFn)
}
```

return to our `main` function to create an instance of it, and then call the
`Get` method in a loop:

```go
func main() {
	// ...

	cacheClient := sturdyc.New[string](...)

	// Create a new API instance with the cache client.
	api := NewAPI(cacheClient)

	// We are going to retrieve the values every 10 milliseconds, however the
	// logs will reveal that actual refreshes fluctuate randomly within a 10-30
	// millisecond range.
	for i := 0; i < 100; i++ {
		val, err := api.Get(context.Background(), "key")
		if err != nil {
			log.Println("Failed to  retrieve the record from the cache.")
			continue
		}
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

This is going to reduce your response times significantly because none of your
users will have to wait for the I/O operation that refreshes the data. It's
always performed in the background as long as the key is being continuously
requested. Being afraid that the record might get too stale if users stop
requesting it is an indication of a TTL that is set too high. Remember, even if
the TTL is exceeded and the key expires, you'll still get deduplication if it's
suddenly requested in a burst again. The only difference is that the users will
have to wait for the I/O operation that retrieves it.

Additionally, to provide a degraded experience when an upstream system
encounters issues, you can set a high TTL and a low refresh time. When
everything is working as expected, the records will be refreshed continuously.
However, if the upstream system encounters issues and stops responding, you can
fall back to cached records for the duration of the TTL.

What if the record was deleted? Our cache might use a 2-hour-long TTL, and we
definitely don't want it to take that long for the deletion to propagate.

However, if we were to modify our client so that it returns an error after the
first request:

```go
type API struct {
	count int
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{0, c}
}

func (a *API) Get(ctx context.Context, key string) (string, error) {
	fetchFn := func(_ context.Context) (string, error) {
		a.count++
		log.Printf("Fetching value for key: %s\n", key)
		if a.count == 1 {
			return "value", nil
		}
		return "", errors.New("error this key does not exist")
	}
	return a.GetOrFetch(ctx, key, fetchFn)
}
```

and then run the program again:

```sh
cd examples/stampede
go run .
```

We'll see that the exponential backoff kicks in, resulting in more iterations
for every refresh, but the value is still being printed:

```sh
2024/05/09 13:22:03 Fetching value for key: key
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:03 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Value: value
2024/05/09 13:22:04 Fetching value for key: key
```

This is a bit tricky because how you determine if a record has been deleted is
going to vary based on your data source. It could be a status code, zero value,
empty list, specific error message, etc. There is no way for the cache to
figure this out implicitly.

It couldn't simply delete a record every time it receives an error. If an
upstream system goes down, we want to be able to serve stale data for the
duration of the TTL, while reducing the frequency of our refreshes to make it
easier for them to recover.

Therefore, if a record is deleted, we'll have to explicitly inform the cache
about it by returning a custom error:

```go
fetchFn := func(_ context.Context) (string, error) {
		a.count++
		log.Printf("Fetching value for key: %s\n", key)
		if a.count == 1 {
			return "value", nil
		}
		return "", sturdyc.ErrNotFound
	}
```

This tell's the cache that the record is no longer available at the underlying data source.
Therefore, if this record is being fetched as a background refresh, the cache will quickly see
if it has a record for this key, and subsequently delete it.

If we run this application again we'll see that it works, and that we're no
longer getting any cache hits. This leads to outgoing requests for every
iteration:

```go
2024/05/09 13:40:47 Fetching value for key: key
2024/05/09 13:40:47 Value: value
2024/05/09 13:40:47 Value: value
2024/05/09 13:40:47 Value: value
2024/05/09 13:40:47 Fetching value for key: key
2024/05/09 13:40:47 Failed to  retrieve the record from the cache.
2024/05/09 13:40:47 Fetching value for key: key
2024/05/09 13:40:47 Failed to  retrieve the record from the cache.
2024/05/09 13:40:47 Fetching value for key: key
2024/05/09 13:40:47 Failed to  retrieve the record from the cache.
2024/05/09 13:40:47 Fetching value for key: key
2024/05/09 13:40:47 Failed to  retrieve the record from the cache.
```

**Please note** that we only have to return the `sturdyc.ErrNotFound` when
we're using `GetOrFetch`. For `GetOrFetchBatch`, we'll simply omit the key from the
map we're returning. I think this inconsistency is a little unfortunate, but it
was the best API I could come up with. Having to return an error like this if
just a single ID wasn't found:

```go
	batchFetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		response, err := myDataSource(cacheMisses)
		for _, id := range cacheMisses {
			// NOTE: Don't do this, it's just an example.
			if response[id]; !id {
                return response, sturdyc.ErrNotFound
            }
		}
		return response, nil
	}
```

and then have the cache swallow that error and return nil, felt much less
intuitive.

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/refreshes)

# Non-existent records

In the example above, we could see that once we delete the key, the following
iterations lead to a continuous stream of outgoing requests. This will happen
for every ID that doesn't exist at the data source. If we can't retrieve it, we
can't cache it. If we can't cache it, we can't serve it from memory. If this
happens frequently, we'll experience a lot of I/O operations, which will
significantly increase our system's latency.

The reasons why someone might request IDs that don't exist can vary. It could
be due to a faulty CMS configuration, or perhaps it's caused by a slow
ingestion process where it takes time for a new entity to propagate through a
distributed system. Regardless, this will negatively impact our systems
performance.

To address this issue, we can instruct the cache to mark these IDs as missing
records. Missing records are refreshed at the same frequency as regular
records. Hence, if an ID is continuously requested, and the upstream eventually
returns a valid response, we'll see it propagate to our cache.

To illustrate, I'll make some small modifications to the code from the previous
example. The only thing I'm going to change is to make the API client return a
`ErrNotFound` for the first three requests:

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
		a.count++
		log.Printf("Fetching value for key: %s\n", key)
		if a.count > 3 {
			return "value", nil
		}
		// This error tells the cache that the data does not exist at the source.
		return "", sturdyc.ErrStoreMissingRecord
	}
	return a.GetOrFetch(ctx, key, fetchFn)
}
```

Next, we'll just have to enable missing record storage which tells the cache
that anytime it gets a `ErrNotFound` error it should mark the key as missing:

```go
func main() {
	// ...

	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithMissingRecordStorage(),
	)

	api := NewAPI(cacheClient)

	// ...
	for i := 0; i < 100; i++ {
		val, err := api.Get(context.Background(), "key")
		// The cache returns ErrMissingRecord for any key that has been marked as missing.
		// You can use this to exit-early, or return some type of default state.
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
refreshes, and then transitions into having a value:

```sh
❯ go run .
2024/05/09 21:25:28 Fetching value for key: key
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Fetching value for key: key
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Fetching value for key: key
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Record does not exist.
2024/05/09 21:25:28 Fetching value for key: key
2024/05/09 21:25:28 Value: value
2024/05/09 21:25:28 Value: value
2024/05/09 21:25:28 Value: value
2024/05/09 21:25:28 Fetching value for key: key
...
```

**Please note** that this functionality is _implicit_ for `GetOrFetchBatch`.
You simply just have to omit the key from the map:

```go
	batchFetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		// The cache will check if every ID in cacheMisses is present in the response.
		// If it finds any IDs that are missing it will proceed to mark them as missing
		// if missing record storage is enabled.
		response, err := myDataSource(cacheMisses)
		return response, nil
	}
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/missing)

# Batch endpoints

One challenge with caching batchable endpoints is that you have to find a way
to reduce the number of keys. To illustrate, let's say that we have 10 000
records, and an endpoint for fetching them that allows for batches of 20.
The IDs for the batch are supplied as query parameters, for example,
`https://example.com?ids=1,2,3,4,5,...20`. If we were to use this as the cache
key, the way many CDNs would do, we could quickly calculate the number of keys
we would generate like this:

$$ C(n, k) = \binom{n}{k} = \frac{n!}{k!(n-k)!} $$

For $n = 10,000$ and $k = 20$, this becomes:

$$ C(10,000, 20) = \binom{10,000}{20} = \frac{10,000!}{20!(10,000-20)!} $$

This results in an approximate value of:

$$ \approx 4.032 \times 10^{61} $$

and this is if we're sending perfect batches of 20. If we were to do 1 to 20
IDs (not just exactly 20 each time) the total number of combinations would be
the sum of combinations for each k from 1 to 20.

At this point, we would essentially just be paying for extra RAM, as the hit
rate for each key would be so low that we'd have better odds of winning the
lottery.

To prevent this, `sturdyc` pulls the response apart and caches each record
individually. This effectively prevents super-polynomial growth in the number
of cache keys because the batch itself is never going to be inlcuded in the
key.

To get a feeling for how this works, let's once again build a small example
application. This time, we'll start with the API client:

```go
type API struct {
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c}
}

func (a *API) GetBatch(ctx context.Context, ids []string) (map[string]string, error) {
	// We are going to use a cache a key function that prefixes each id.
	// This makes it possible to save the same id for different data sources.
	cacheKeyFn := a.BatchKeyFn("some-prefix")

	// The fetchFn is only going to retrieve the IDs that are not in the cache.
	fetchFn := func(_ context.Context, cacheMisses []string) (map[string]string, error) {
		log.Printf("Cache miss. Fetching ids: %s\n", strings.Join(cacheMisses, ", "))
		// Batch functions should return a map where the key is the id of the record.
		response := make(map[string]string, len(cacheMisses))
		for _, id := range cacheMisses {
			response[id] = "value"
		}
		return response, nil
	}

	return a.GetOrFetchBatch(ctx, ids, cacheKeyFn, fetchFn)
}
```

and we're going to use the same cache configuration as the previous example, so
I've omitted it for brevity:

```go
func main() {
	// ...

	// Create a new API instance with the cache client.
	api := NewAPI(cacheClient)

	// Make an initial call to make sure that IDs 1-10 are retrieved and cached.
	log.Println("Seeding ids 1-10")
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	api.GetBatch(context.Background(), ids)
	log.Println("Seed completed")

	// To demonstrate that the records have been cached individually, we can continue
	// fetching a random subset of records from the original batch, plus a new
	// ID. By examining the logs, we should be able to see that the cache only
	// fetches the ID that wasn't present in the original batch, indicating that
	// the batch itself isn't part of the key.
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
ways.

Consider this:

```sh
curl https://movie-api/movies?ids=1,2,3&filterUpcoming=true&includeTrailers=false
curl https://movie-api/movies?ids=1,2,3&filterUpcoming=false&includeTrailers=true
```

The IDs might be enough to uniquely identify these records in a database.
However, when you're consuming them through another system, they will probably
appear completely different as transformations are applied based on the options
you pass it. Hence, it's important that we store these records once for each
unique option set.

The options does not have to be query parameters either. The data source you're
consuming could still be a database, and the options that you want to make part
of the cache key could be different types of filters.

Below is a small example application to showcase this functionality:

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
	// We use the PermutedBatchKeyFn when an ID isn't enough to uniquely identify a
	// record. The cache is going to store each id once per set of options.
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
	return a.GetOrFetchBatch(ctx, ids, cacheKeyFn, fetchFn)
}
```

The main difference from the previous example is that we're using
`PermutatedBatchKeyFn` instead of `BatchKeyFn`. Internally, the cache will use
reflection to extract the names and values of every **exported** field in the
`opts` struct, and then include them when it constructs the cache keys.

The struct should be flat without nesting. The fields can be `time.Time`
values, as well as any basic types, pointers to these types, and slices
containing them.

Now, let's try to use this client:

```go
func main() {
	// ...

	// Create a new cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
	)

	// We will fetch these IDs using three different option sets.
	ids := []string{"id1", "id2", "id3"}
	optionSetOne := OrderOptions{CarrierName: "FEDEX", LatestDeliveryTime: "2024-04-06"}
	optionSetTwo := OrderOptions{CarrierName: "DHL", LatestDeliveryTime: "2024-04-07"}
	optionSetThree := OrderOptions{CarrierName: "UPS", LatestDeliveryTime: "2024-04-08"}

	orderClient := NewOrderAPI(cacheClient)
	ctx := context.Background()

	// Next, we'll call the orderClient to make sure that we've retrieved and cached
	// these IDs for all of our option sets.
	log.Println("Filling the cache with all IDs for all option sets")
	orderClient.OrderStatus(ctx, ids, optionSetOne)
	orderClient.OrderStatus(ctx, ids, optionSetTwo)
	orderClient.OrderStatus(ctx, ids, optionSetThree)
	log.Println("Cache filled")
}
```

At this point, the cache has stored each record individually for each option
set. We can imagine that the keys would look something like this:

```
FEDEX-2024-04-06-id1
DHL-2024-04-07-id1
UPS-2024-04-08-id1
etc..
```

Next, we'll add a sleep to make sure that all of the records are due for a
refresh, and then request the ids individually for each set of options:

```go
func main() {
	// ...

	// Sleep to make sure that all records are due for a refresh.
	time.Sleep(maxRefreshDelay + 1)

	// Fetch each id for each option set.
	for i := 0; i < len(ids); i++ {
		// NOTE: We're using the same ID for these requests.
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
2024/04/07 13:33:58 Fetching: [id1], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id1], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id1], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:33:58 Fetching: [id2], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id2], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id2], carrier: DHL, delivery time: 2024-04-07
2024/04/07 13:33:58 Fetching: [id3], carrier: FEDEX, delivery time: 2024-04-06
2024/04/07 13:33:58 Fetching: [id3], carrier: UPS, delivery time: 2024-04-08
2024/04/07 13:33:58 Fetching: [id3], carrier: DHL, delivery time: 2024-04-07
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/permutations)

# Refresh coalescing

As seen in the example above, we're storing the records once for every set of
options. However, we're not really utilizing the fact that the endpoint is
batchable when we're performing the refreshes.

To make this more efficient, we can enable the **refresh coalescing**
functionality. Internally, the cache is going to create a buffer for every
cache key permutation. It is then going to collect ids until it reaches a
certain size, or exceeds a time-based threshold.

The only change we have to make to the previous example is to enable this
feature:

```go
func main() {
	// ...

	// With refresh coalescing enabled, the cache will buffer refreshes
	// until the batch size is reached or the buffer timeout is hit.
	batchSize := 3
	batchBufferTimeout := time.Second * 30

	// Create a new cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithRefreshCoalescing(batchSize, batchBufferTimeout),
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

The number of outgoing requests for the refreshes went from **9** to **3**.
Imagine what a batch size of 50 would do for your applications performance!

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/buffering)

# Passthrough

There are times when you want to always retrieve the latest data from the
source and only use the in-memory cache as a fallback. In such scenarios, you
can use the `Passthrough` and `PassthroughBatch` functions. The cache will
still perform in-flight request tracking and deduplicate your requests.

# Distributed storage

I think it's important to read the previous sections before jumping here in
order to understand all the heavy lifting `sturdyc` does when it comes to
creating cache keys, tracking in-flight requests, refreshing records in the
background to improve latency, and buffering/coalescing requests to minimize
the number of round trips to underlying data sources.

Adding distributed storage to the cache is, from the package's point of view,
essentially just another data source with a higher priority. Hence, we're still
able to take great advantage of all the features we've seen so far, and these
efficiency gains will hopefully allow you to use a much cheaper cluster.

Slightly simplified, we can think of the cache's interaction with the
distributed storage like this:

```go
// NOTE: This is an example. The cache has this functionality internally.
func (o *OrderAPI) OrderStatus(ctx context.Context, id string) (string, error) {
	cacheKey := "order-status-" + id
	fetchFn := func(ctx context.Context) (string, error) {
		// Check redis cache first.
		if orderStatus, ok := o.redisClient.Get(cacheKey); ok {
			return orderStatus, nil
		}

		// Fetch the order status from the underlying data source.
		var response OrderStatusResponse
		err := requests.URL(o.baseURL).
			Param("id", id).
			ToJSON(&response).
			Fetch(ctx)
		if err != nil {
			return "", err
		}

		// Add the order status to the redis cache.
		go func() { o.RedisClient.Set(cacheKey, response.OrderStatus, time.Hour) }()

		return response.OrderStatus, nil
	}

	return o.GetOrFetch(ctx, id, fetchFn)
}
```

Syncing the keys and values to a distributed storage like this can be highly
beneficial, especially when we're deploying new containers where the in-memory
cache will be empty, as it prevents sudden bursts of traffic to the underlying
data sources.

Keeping the in-memory caches in sync with a distributed storage requires a bit
more work though. `sturdyc` has therefore been designed to work with an
abstraction that could represent any key-value store of your choosing, all you
have to do is implement this interface:

```go
type DistributedStorage interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte)
	GetBatch(ctx context.Context, keys []string) map[string][]byte
	SetBatch(ctx context.Context, records map[string][]byte)
}
```

and then pass it to the `WithDistributedStorage` option when you create your
cache client:

```go
cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
	sturdyc.WithDistributedStorage(storage),
)
```

**Please note** that you are responsible for configuring the TTL and eviction
policies of this storage. `sturdyc` will only make sure that it's being kept
up-to-date with the data it has in-memory.

I've included an example to showcase this functionality
[here.](https://github.com/creativecreature/sturdyc/tree/main/examples/distribution)

When running that application, you should see output that looks something like
this:

```go
❯ go run .
2024/06/07 10:32:56 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:56 Fetching shipping options from the underlying data source
2024/06/07 10:32:56 The shipping options were retrieved successfully!
2024/06/07 10:32:56 Writing key shipping-options-1234-asc to the distributed storage
2024/06/07 10:32:56 The shipping options were retrieved successfully!
2024/06/07 10:32:57 The shipping options were retrieved successfully!
2024/06/07 10:32:57 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:57 The shipping options were retrieved successfully!
2024/06/07 10:32:57 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:57 The shipping options were retrieved successfully!
2024/06/07 10:32:57 The shipping options were retrieved successfully!
2024/06/07 10:32:57 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:58 The shipping options were retrieved successfully!
2024/06/07 10:32:58 The shipping options were retrieved successfully!
2024/06/07 10:32:58 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:58 The shipping options were retrieved successfully!
2024/06/07 10:32:58 Getting key shipping-options-1234-asc from the distributed storage
2024/06/07 10:32:58 The shipping options were retrieved successfully!
```

Above we can see that the underlying data source was only visited **once**, and
the in-memory cache performed a background refresh from the distributed storage
every 2 to 3 retrievals to ensure that it's being kept up-to-date.

This sequence of events will repeat once the TTL expires.

# Distributed storage early refreshes

Similar to the in-memory cache, we're also able to use a distributed storage
where the data is refreshed before the TTL expires.

This would also allow us to serve stale data if an upstream was to experience
any downtime:

```go
cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
	sturdyc.WithDistributedStorageEarlyRefreshes(storage, time.Minute),
)
```

With the configuration above, we're essentially saying that we'd prefer if the
data was refreshed once it's more than a minute old. However, if you're writing
records with a 60 minute TTL, the cache will continously fallback to these if
the refreshes were to fail, so the interaction with the distributed storage
would look something like this:

- Start by trying to retrieve the key from the distributeted storage. If the
  data is fresh, it's returned immediately and written to the in-memory cache.
- If the key was found in the distributed storage, but wasn't fresh enough,
  we'll visit the underlying data source, and then write the response to both
  the distributed cache and the one we have in-memory.
- If the call to refresh the data failed, the cache will use the value from the
  distributed storage as a fallback.

However, there is one more scenario we must cover that requires two additional
methods to be implemented:

```go
type DistributedStorageEarlyRefreshes interface {
	DistributedStorage
	Delete(ctx context.Context, key string)
	DeleteBatch(ctx context.Context, keys []string)
}
```

These delete methods will be called when a refresh occurs, and the cache
notices that it can no longer find the key at the underlying data source. This
indicates that the key has been deleted, and we will want this change to
propagate to the distributed key-value store

**Please note** that you are still responsible for setting the TTL and eviction
policies for the distributed store. The cache will only invoke the delete
methods when a record has gone missing from the underlying data source. If
you're using **missing record storage**, it will write the key as a missing
record instead.

I've included an example to showcase this functionality
[here.](https://github.com/creativecreature/sturdyc/tree/main/examples/distributed-early-refreshes)

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

There are also distributed metrics if you're using the cache with a
_distributed storage_, which adds the following metrics in addition to what
we've seen above:

- Distributed cache hits
- Distributed cache misses
- Distributed stale fallback

All you have to do is implement one of these interfaces:

```go
type MetricsRecorder interface {
	CacheMiss()
	Eviction()
	ForcedEviction()
	EntriesEvicted(int)
	ShardIndex(int)
	CacheBatchRefreshSize(size int)
	ObserveCacheSize(callback func() int)
}

type DistributedMetricsRecorder interface {
	MetricsRecorder
	DistributedCacheHit()
	DistributedCacheMiss()
	DistributedFallback()
}

```

and pass it as an option when you create the client:

```go
cacheBasicMetrics := sturdyc.New[any](
	cacheSize,
	shardSize,
	cacheTTL,
	evictWhenFullPercentage,
	sturdyc.WithMetrics(metricsRecorder),
)

cacheDistributedMetrics := sturdyc.New[any](
	cacheSize,
	shardSize,
	cacheTTL,
	evictWhenFullPercentage,
	sturdyc.WithDistributedStorage(metricsRecorder),
	sturdyc.WithDistributedMetrics(metricsRecorder),
)
```

Below are a few images where these metrics have been visualized in Grafana:

<img width="939" alt="Screenshot 2024-05-04 at 12 36 43" src="https://github.com/creativecreature/sturdyc/assets/12787673/1f630aed-2322-4d3a-9510-d582e0294488">
Here we can how often we're able to serve from memory.

<img width="942" alt="Screenshot 2024-05-04 at 12 37 39" src="https://github.com/creativecreature/sturdyc/assets/12787673/25187529-28fb-4c4e-8fe9-9fb48772e0c0">
This image displays the number of items we have cached.

<img width="941" alt="Screenshot 2024-05-04 at 12 38 04" src="https://github.com/creativecreature/sturdyc/assets/12787673/b1359867-f1ef-4a09-8c75-d7d2360726f1">
This chart shows the batch sizes for the buffered refreshes.

<img width="940" alt="Screenshot 2024-05-04 at 12 38 20" src="https://github.com/creativecreature/sturdyc/assets/12787673/de7f00ee-b14d-443b-b69e-91e19665c252">
And lastly, we can see the average batch size of our refreshes for two different data sources.

You are also able to visualize evictions, forced evictions which occur when the
cache has reached its capacity, as well as the distribution between the shards.

# Generics

Personally, I tend to create caches based on how frequently the data needs to
be refreshed rather than what type of data it stores. I'll often have one
transient cache which refreshes the data every 2-5 milliseconds, and another
cache where I'm fine if the data is up to a minute old.

Hence, I don't want to tie the cache to any specific type so I'll often just
use `any`:

```go
	cacheClient := sturdyc.New[any](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
		sturdyc.WithRefreshCoalescing(10, time.Second*15),
	)
```

However, having all client methods return `any` can quickly add a lot of
boilerplate if you're storing more than a handful of types, and need to make
type assertions.

If you want to avoid this, you can use any of the package level exports:

- [`GetOrFetch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#GetOrFetch)
- [`GetOrFetchBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#GetOrFetchBatch)
- [`Passthrough`](https://pkg.go.dev/github.com/creativecreature/sturdyc#Passthrough)
- [`PassthroughBatch`](https://pkg.go.dev/github.com/creativecreature/sturdyc#PassthroughBatch)

They will take the cache, call the function for you, and perform the type
conversions internally. If the type conversions were to fail, you'll get a
[`ErrInvalidType`](https://pkg.go.dev/github.com/creativecreature/sturdyc#pkg-variables) error.

Below is an example of what an API client that uses these functions could look
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
	return sturdyc.GetOrFetchBatch(ctx, a.cacheClient, ids, cacheKeyFn, fetchFn)
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
	return sturdyc.GetOrFetchBatch(ctx, a.cacheClient, ids, cacheKeyFn, fetchFn)
}
```

The entire example is available [here.](https://github.com/creativecreature/sturdyc/tree/main/examples/generics)
