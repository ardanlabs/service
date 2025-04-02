![sturdyC-fn-2](https://github.com/user-attachments/assets/2def120a-ad2b-4590-bef0-83c461af1b07)
> A sturdy gopher shielding data sources from rapidly incoming requests.

# `sturdyc`: a caching library for building sturdy systems

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/viccon/sturdyc.svg)](https://pkg.go.dev/github.com/viccon/sturdyc)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/viccon/sturdyc/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/viccon/sturdyc)](https://goreportcard.com/report/github.com/viccon/sturdyc)
[![Test](https://github.com/viccon/sturdyc/actions/workflows/main.yml/badge.svg)](https://github.com/viccon/sturdyc/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/viccon/sturdyc/graph/badge.svg?token=CYSKW3Z7E6)](https://codecov.io/gh/viccon/sturdyc)


`sturdyc` eliminates cache stampedes and can minimize data source load in
high-throughput systems through features such as request coalescing and
asynchronous refreshes. It combines the speed of in-memory caching with
granular control over data freshness. At its core, `sturdyc` provides
**non-blocking reads** and **sharded writes** for minimal lock contention. The
[xxhash](https://github.com/cespare/xxhash) algorithm is used for efficient key
distribution.

It has all the functionality you would expect from a caching library, but what
**sets it apart** are the flexible configurations that have been designed to
make I/O heavy applications both _robust_ and _highly performant_.

We have been using this package in production to enhance both the performance
and reliability of our services that retrieve data from distributed caches,
databases, and external APIs. While the API surface of sturdyc is tiny, it
offers extensive configuration options. I encourage you to read through this
README and experiment with the examples in order to understand its full
capabilities.

This screenshot shows the P95 latency improvements we observed after adding
this package in front of a distributed key-value store:

&nbsp;
<img width="1554" alt="Screenshot 2024-05-10 at 10 15 18" src="https://github.com/viccon/sturdyc/assets/12787673/adad1d4c-e966-4db1-969a-eda4fd75653a">
&nbsp;

And through a combination of inflight-tracking, asynchronous refreshes, and
refresh coalescing, we reduced load on underlying data sources by more than
90%. This reduction in outgoing requests has enabled us to operate with fewer
containers and significantly cheaper database clusters.

# Table of contents

Below is the table of contents for what this README is going to cover. However,
if this is your first time using this package, I encourage you to **read these
examples in the order they appear**. Most of them build on each other, and many
share configurations.

- [**installing**](https://github.com/viccon/sturdyc?tab=readme-ov-file#installing)
- [**creating a cache client**](https://github.com/viccon/sturdyc?tab=readme-ov-file#creating-a-cache-client)
- [**evictions**](https://github.com/viccon/sturdyc?tab=readme-ov-file#evictions)
- [**get or fetch**](https://github.com/viccon/sturdyc?tab=readme-ov-file#get-or-fetch)
- [**stampede protection**](https://github.com/viccon/sturdyc?tab=readme-ov-file#stampede-protection)
- [**early refreshes**](https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes)
- [**deletions**](https://github.com/viccon/sturdyc?tab=readme-ov-file#deletions)
- [**caching non-existent records**](https://github.com/viccon/sturdyc?tab=readme-ov-file#non-existent-records)
- [**caching batch endpoints per record**](https://github.com/viccon/sturdyc?tab=readme-ov-file#batch-endpoints)
- [**cache key permutations**](https://github.com/viccon/sturdyc?tab=readme-ov-file#cache-key-permutations)
- [**refresh coalescing**](https://github.com/viccon/sturdyc?tab=readme-ov-file#refresh-coalescing)
- [**request passthrough**](https://github.com/viccon/sturdyc?tab=readme-ov-file#passthrough)
- [**distributed storage**](https://github.com/viccon/sturdyc?tab=readme-ov-file#distributed-storage)
- [**custom metrics**](https://github.com/viccon/sturdyc?tab=readme-ov-file#custom-metrics)
- [**generics**](https://github.com/viccon/sturdyc?tab=readme-ov-file#generics)

# Installing

```sh
go get github.com/viccon/sturdyc
```

# Creating a cache client

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


The `New` function is variadic, and as the final argument we're also able to
provide a wide range of configuration options, which we will explore in detail
in the sections to follow.

# Evictions

The cache has two eviction strategies. One is a background job which
continuously evicts expired records from each shard. However, there are options
to both tweak the interval at which the job runs:

```go
cacheClient := sturdyc.New[int](capacity, numShards, ttl, evictionPercentage,
    sturdyc.WithEvictionInterval(time.Second),
)
```

as well as disabling the functionality altogether:

```go
cacheClient := sturdyc.New[int](capacity, numShards, ttl, evictionPercentage,
    sturdyc.WithNoContinuousEvictions()
)
```

The latter can give you a slight performance boost in situations where you're
unlikely to ever exceed the capacity you've assigned to your cache.

However, if the capacity is reached, the second eviction strategy is triggered.
This process performs evictions on a per-shard basis, selecting records for
removal based on recency. The eviction algorithm uses
[quickselect](https://en.wikipedia.org/wiki/Quickselect), which has an O(N)
time complexity without the overhead of requiring write locks on reads to
update a recency list, as many LRU caches do.

Next, we'll start to look at some of the more _advanced features_.

# Get or fetch

I have tried to design the API in a way that should make the process of
integrating `sturdyc` with any data source as straightforward as possible.
While it provides the basic get/set methods you would expect from a cache, the
advanced functionality is accessed through just two core functions:
`GetOrFetch` and `GetOrFetchBatch`

As an example, let's say that we had the following code for fetching orders
from an API:

```go
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
```

All we would have to do is wrap the lines of code that retrieves the data in a
function, and then hand that over to our cache client:

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

	return c.cache.GetOrFetch(ctx, id, fetchFunc)
}
```

The cache is going to return the value from memory if it's available, and
otherwise will call the `fetchFn` to retrieve the data from the underlying data
source.

Most of our examples are going to be retrieving data from HTTP APIs, but it's
just as easy to wrap a database query, a remote procedure call, a disk read, or
any other I/O operation.

The `fetchFn` that we pass to `GetOrFetch` has the following function
signature:

```go
type FetchFn[T any] func(ctx context.Context) (T, error)
```

For data sources capable of handling requests for multiple records at once,
we'll use `GetOrFetchBatch`:

```go
type KeyFn func(id string) string

type BatchFetchFn[T any] func(ctx context.Context, ids []string) (map[string]T, error)

func (c *Client[T]) GetOrFetchBatch(ctx context.Context, ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) (map[string]T, error) {
	// ...
}
```

There are a few things to unpack here, so let's start with the `KeyFn`. When
adding an in-memory cache to an API client capable of calling multiple
endpoints, it's highly unlikely that an ID alone is going to be enough to
uniquely identify a record.

To illustrate, let's say that we're building a Github client and want to use
this package to get around their rate limit. The username itself wouldn't make
for a good cache key because we could use it to fetch gists, commits,
repositories, etc. Therefore, `GetOrFetchBatch` takes a `KeyFn` that prefixes
each ID with something to identify the data source so that we don't end up with
cache key collisions:

```go
gistPrefixFn := cacheClient.BatchKeyFn("gists")
commitPrefixFn := cacheClient.BatchKeyFn("commits")
gists, err := cacheClient.GetOrFetchBatch(ctx, userIDs, gistPrefixFn, fetchGists)
commits, err := cacheClient.GetOrFetchBatch(ctx, userIDs, commitPrefixFn, fetchCommits)
```

We're now able to use the _same_ cache for _multiple_ data sources, and
internally we'd get cache keys of this format:

```
gists-ID-viccon
gists-ID-some-other-user
commits-ID-viccon
commits-ID-some-other-user
```

Now, let's use a bit of our imagination because Github doesn't actually allow
us to fetch gists from multiple users at once. However, if they did, our client
would probably look something like this:

```go
func (client *GithubClient) Gists(ctx context.Context, usernames []string) (map[string]Gist, error) {
	cacheKeyFn := client.cache.BatchKeyFn("gists")
	fetchFunc := func(ctx context.Context, cacheMisses []string) (map[string]Gist, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, client.timeout)
		defer cancel()

		var response map[string]Gist
		err := requests.URL(c.baseURL).
			Path("/gists").
			Param("usernames", strings.Join(cacheMisses, ",")).
			ToJSON(&response).
			Fetch(timeoutCtx)
		return response, err
	}
	return sturdyc.GetOrFetchBatch(ctx, client.cache, usernames, cacheKeyFn, fetchFunc)
}
```

In the example above, the `fetchFunc` would get called for users where we don't
have their gists in our cache, and the cacheMisses slice would contain their
actual usernames (without the prefix from the keyFn).

The map that we return from our `fetchFunc` should have the IDs (in this case the
usernames) as keys, and the actual data that we want to cache (the gist) as the
value.

Later, we'll see how we can use closures to pass query parameters and options
to our fetch functions, as well as how to use the `PermutatedBatchKeyFn` to
create unique cache keys for each permutation of them.

# Stampede protection

`sturdyc` provides automatic protection against cache stampedes (also known as
thundering herd) - a situation that occurs when many requests for a particular
piece of data, which has just expired or been evicted from the cache, come in
at once.

Preventing this has been one of the key objectives. We do not want to cause a
significant load on an underlying data source every time one of our keys
expire. To address this, `sturdyc` performs _in-flight_ tracking for every key.

We can demonstrate this using the `GetOrFetch` function which, as I mentioned
earlier, takes a key, and a function for retrieving the data if it's not in the
cache. The cache is going to ensure that we never have more than a single
in-flight request per key:

```go
	var count atomic.Int32
	fetchFn := func(_ context.Context) (int, error) {
		// Increment the count so that we can assert how many times this function was called.
		count.Add(1)
		time.Sleep(time.Second)
		return 1337, nil
	}

	// Fetch the same key from 5 goroutines.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			// We'll ignore the error here for brevity.
			val, _ := cacheClient.GetOrFetch(context.Background(), "key2", fetchFn)
			log.Printf("got value: %d\n", val)
			wg.Done()
		}()
	}
	wg.Wait()

	log.Printf("fetchFn was called %d time\n", count.Load())
	log.Println(cacheClient.Get("key2"))

```

Running this program we can see that we were able to retrieve the value for all
5 goroutines, and that the fetchFn only got called once:

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

The in-flight tracking works for batch operations too where the cache is able
to deduplicate a batch of cache misses, and then assemble the response by
picking records from **multiple** in-flight requests.

To demonstrate this, we'll use the `GetOrFetchBatch` function, which as mentioned
earlier, can be used to retrieve data from a data source capable of handling
requests for multiple records at once.

We'll start by creating a mock function that sleeps for `5` seconds, and then
returns a map with a numerical value for every ID:

```go
var count atomic.Int32
fetchFn := func(_ context.Context, ids []string) (map[string]int, error) {
	// Increment the counter so that we can assert how many times this function was called.
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

Next, we'll need some batches to test with, so here I've created three batches
with five IDs each:

```go
batches := [][]string{
	{"1", "2", "3", "4", "5"},
	{"6", "7", "8", "9", "10"},
	{"11", "12", "13", "14", "15"},
}
```

and we can now request each batch in a separate goroutine:

```go
for _, batch := range batches {
	go func() {
		res, _ := cacheClient.GetOrFetchBatch(context.Background(), batch, keyPrefixFn, fetchFn)
		log.Printf("got batch: %v\n", res)
	}()
}

// Just to ensure that these batches are in fact in-flight, we'll sleep to give the goroutines a chance to run.
time.Sleep(time.Second * 2)
```

At this point, the cache should have 3 in-flight requests for IDs 1-15:

```sh
[1,2,3,4,5]      => REQUEST 1 (IN-FLIGHT)
[6,7,8,9,10]     => REQUEST 2 (IN-FLIGHT)
[11,12,13,14,15] => REQUEST 3 (IN-FLIGHT)
```

Knowing this, let's test the stampede protection by launching another five
goroutines. Each of these goroutines will request two random IDs from our
previous batches. For example, they could request one ID from the first
request, and another from the second or third.

```go
// Launch another 5 goroutines that are going to pick two random IDs from any of our in-flight batches.
// e.g:
// [1,8]
// [4,11]
// [14,2]
// [6,15]

func pickRandomValue(batches [][]string) string {
	batch := batches[rand.IntN(len(batches))]
	return batch[rand.IntN(len(batch))]
}

var wg sync.WaitGroup
for i := 0; i < 5; i++ {
	wg.Add(1)
	go func() {
		ids := []string{pickRandomValue(batches), pickRandomValue(batches)}
		res, _ := cacheClient.GetOrFetchBatch(context.Background(), ids, keyPrefixFn, fetchFn)
		log.Printf("got batch: %v\n", res)
		wg.Done()
	}()
}

wg.Wait()
log.Printf("fetchFn was called %d times\n", count.Load())
```

Running this program, and looking at the logs, we'll see that the cache is able
to resolve all of the ids from these new goroutines without generating any
additional requests even though we're picking IDs from different in-flight
requests:

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

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/basic)

# Early refreshes

Serving data from memory is typically at least one to two orders of magnitude
faster than reading from disk, and if you have to retrieve the data across a
network the difference can grow even larger. Consequently, we're often able to
significantly improve our applications performance by adding an in-memory
cache.

However, one has to be aware of the usual trade-offs. Suppose we use a TTL of
10 seconds. That means the cached data can be up to 10 seconds old. In many
applications this may be acceptable, but in others it can introduce stale
reads. Additionally, once the cached value expires, the first request after
expiration must refresh the cache, resulting in a longer response time for that
user. This can make the average latency look very different from the P90–P99
tail latencies, since those percentiles capture the delays of having to go to
the actual data source in order to refresh the cache. This in turn can make it
difficult to configure appropriate alarms for your applications response times.

`sturdyc` aims to give you a lot of control over these choices when you enable
the **early refreshes** functionality. It will prevent your most frequently
used records from ever expiring by continuously refreshing them in the
background. This can have a significant impact on your applications latency.
We've seen the P99 of some of our applications go from 50ms down to 1.

One thing to note about these background refreshes is that they are scheduled
if a key is **requested again** after a configurable amount of time has passed.
This is an important distinction because it means that the cache doesn't just
naively refresh every key it's ever seen. Instead, it only refreshes the
records that are actually in active rotation, while allowing unused keys to be
deleted once their TTL expires. This also means that the request that gets
chosen to refresh the value won’t retrieve the updated data right away as the
refresh happens asynchronously.

However, asynchronous refreshes present challenges with infrequently requested
keys. While background refreshes keep latency low by serving cached values
during updates, this can lead to perpetually stale data. If a key isn't
requested again before its next scheduled refresh, we remain permanently one
update behind, as each read triggers a refresh that won't be seen until the
next request. This is similar to a burger restaurant that prepares a new burger
after each customer's order - if the next customer arrives too late, they'll
receive a cold burger, despite the restaurant's proactive cooking strategy.

To solve this, you also get to provide a synchronous refresh time. This
essentially tells the cache: "If the data is older than x, I want the refresh
to be blocking and have the user wait for the response." Or using the burger
analogy: if a burger has been sitting for more than X minutes, the restaurant
starts making a fresh one while the customer waits. Unlike a real restaurant
though, the cache keeps the old value as a fallback - if the refresh fails,
we'll still serve the "cold burger" rather than letting the customer go hungry.

Below is an example configuration that you can use to enable this
functionality:

```go
func main() {
	// Set a minimum and maximum refresh delay for the records. This is used to
	// spread out the refreshes of our records evenly over time. If we're running
	// our application across 100 containers, we don't want to send a spike of
	// refreshes from every container every 30 ms. Instead, we'll use some
	// randomization to spread them out evenly between 10 and 30 ms.
	minRefreshDelay := time.Millisecond * 10
	maxRefreshDelay := time.Millisecond * 30
	// Set a synchronous refresh delay for when we want a refresh to happen synchronously.
	synchronousRefreshDelay := time.Second * 30
	// The base used for exponential backoff when retrying a background refresh.
	// Most of the time, we perform refreshes well in advance of the records
	// expiry time. Hence, we can use this to make it easier for a system that
	// is having trouble to get back on it's feet by making fewer refreshes when
	// we're seeing a lot of errors. Once we receive a successful response, the
	// refreshes return to their original frequency. You can set this to 0
	// if you don't want this behavior.
	retryBaseDelay := time.Millisecond * 10

	// Create a cache client with the specified configuration.
	cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
		sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, retryBaseDelay),
	)
}
```

Let's build a simple API client that embeds the cache using our configuration:

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

Now we can return to our `main` function to create an instance of it, and then
call the `Get` method in a loop:

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

Running this program, we're going to see that the value gets refreshed
asynchronously once every 2-3 retrievals:

```sh
cd examples/refreshes
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

If this was a real application it would have reduced our response times
significantly because none of our users would have to wait for the I/O
operation that refreshes the data. It's always performed in the background as
long as the key is being continuously requested.

We don't have to be afraid that the data for infrequently used keys gets stale
either, given that we set the synchronous refresh delay like this:

```go
	synchronousRefreshDelay := time.Second * 30
```

Which means that if a key isn't requested again within 30 seconds, the cache
will make the refresh synchronous. Even if a minute has passed and 1,000
requests suddenly come in for this key, the stampede protection will kick in
and make the refresh synchronous for all of them, while also ensuring that only
a single request is made to the underlying data source.

Sometimes I like to use this feature to provide a degraded experience when an
upstream system encounters issues. For this, I choose a high TTL and a low
refresh time, so that when everything is working as expected, the records are
refreshed continuously. However, if the upstream system stops responding, I can
rely on cached records for the entire duration of the TTL.

This also brings us to the final argument of the `WithEarlyRefreshes` function
which is the retry base delay. This delay is used to create an exponential
backoff for our background requests if a data source starts to return errors.
Please note that this **only** applies to background refreshes. If we reach a
point where all of the records are older than the synchronous refresh time,
we're going to send a steady stream of outgoing requests. That is because I
think of the synchronous refresh time as "I really don’t want the data to be
older than this, but I want the possibility of using an even higher TTL in
order to serve stale." Therefore, if a synchronous refresh fails, I want the
very next request for that key to attempt another refresh.

Also, if you don't want any of this serve stale functionality you could just
use short TTLs. The cache will never return a record where the TTL has expired.
I'm just trying to showcase some different ways to leverage this functionality!

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/refreshes)

# Deletions

What if a record gets deleted at the underlying data source? Our cache might
use a 2-hour-long TTL, and we definitely don't want it to take that long for
the deletion to propagate.

However, if we were to modify our client from the previous example so that it
returns an error after the first request:

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
cd examples/refreshes
go run .
```

We'll see that the exponential backoff kicks in, delaying our background
refreshes which results in more iterations for every refresh, but the value is
still being printed:

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

This is a bit tricky because how you determine if a record has been deleted
could vary based on your data source. It could be a status code, zero value,
empty list, specific error message, etc. There is no easy way for the cache to
figure this out implicitly.

It couldn't simply delete a record every time it receives an error. If an
upstream system goes down, we want to be able to serve the data for the
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

This tells the cache that the record is no longer available at the underlying data source.

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

and then have the cache either swallow that error and return nil, or return the
map with the error, felt much less intuitive.

This code is based on the example available [here.](https://github.com/viccon/sturdyc/tree/main/examples/refreshes)

# Non-existent records

In the example above, we could see that once we delete the key, the following
iterations lead to a continuous stream of outgoing requests. This would also
happen for every ID that doesn't exist at the underlying data source. If we
can't retrieve it, we can't cache it. If we can't cache it, we can't serve it
from memory. If this happens frequently, we'll experience a lot of I/O
operations, which will significantly increase our system's latency.

The reasons why someone might request IDs that don't exist can vary. It could
be due to a faulty CMS configuration, or perhaps it's caused by a slow
ingestion process where it takes time for a new entity to propagate through a
distributed system. Regardless, this will negatively impact our systems
performance.

To address this issue, we can instruct the cache to mark these IDs as missing
records. If you're using this functionality in combination with the
`WithEarlyRefreshes` option, they are going to get refreshed at the same
frequency as regular records. Hence, if an ID is continuously requested, and
the upstream eventually returns a valid response, we'll see it propagate to our
cache.

To illustrate, I'll make some small modifications to the code from the previous
example. I'm going to to make the API client return a `ErrNotFound` for the
first three requests:

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
		return "", sturdyc.ErrNotFound
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
2024/05/09 21:25:28 Value: value // Look, the value exists now!
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

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/missing)

# Batch endpoints

One challenge with caching batchable endpoints is that you have to find a way
to reduce the number of cache keys. Consider an endpoint that allows fetching
10,000 records in batches of 20. The IDs for the batch are supplied as query
parameters, for example, `https://example.com?ids=1,2,3,4,5,...20`. If we were
to use this as the cache key, the way many CDNs would, we could quickly
calculate the number of keys we would generate like this:

$$ C(n, k) = \binom{n}{k} = \frac{n!}{k!(n-k)!} $$

For $n = 10,000$ and $k = 20$, this becomes:

$$ C(10,000, 20) = \binom{10,000}{20} = \frac{10,000!}{20!(10,000-20)!} $$

This results in an approximate value of:

$$ \approx 4.032 \times 10^{61} $$

and this is if we're sending perfect batches of 20. If we were to do 1 to 20
IDs (not just exactly 20 each time) the total number of combinations would be
the sum of combinations for each k from 1 to 20.

At this point, the hit rate for each key would be so low that we'd have better
odds of winning the lottery.

To prevent this, `sturdyc` pulls the response apart and caches each record
individually. This effectively prevents super-polynomial growth in the number
of cache keys because the batch itself is never going to be included in the
key.

To get a better feeling for how this works, we can look at the function signature
for the `GetOrFetchBatch` function:

```go
func (c *Client[T]) GetOrFetchBatch(ctx context.Context, ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) (map[string]T, error) {}
```

What the cache does is that it takes the IDs, applies the `keyFn` to them, and
then checks each key individually if it's present in the cache. The keys that
aren't present will be passed to the `fetchFn`.

The `fetchFn` has this signature where it returns a map where the ID is the
key:

```go
type BatchFetchFn[T any] func(ctx context.Context, ids []string) (map[string]T, error)
```

The cache can use this to iterate through the response map, again apply the
`keyFn` to each ID, and then store each record individually.

Sometimes, the function signature for the `BatchFetchFn` can feel too limited.
You may need additional options and not just the IDs to retrieve the data. But
don't worry, we'll look at how to solve this in the next section!

For now, to get some code to play around with, let's once again build a small
example application. This time, we'll start with the API client:

```go
type API struct {
	*sturdyc.Client[string]
}

func NewAPI(c *sturdyc.Client[string]) *API {
	return &API{c}
}

func (a *API) GetBatch(ctx context.Context, ids []string) (map[string]string, error) {
	// We are going to pass a cache a key function that prefixes each id with
	// the string "some-prefix", and adds an -ID- separator before the actual
	// id. This makes it possible to save the same id for different data
	// sources as the keys would look something like this: some-prefix-ID-1234
	cacheKeyFn := a.BatchKeyFn("some-prefix")

	// The fetchFn is only going to retrieve the IDs that are not in the cache. Please
	// note that the cacheMisses is going to contain the actual IDs, not the cache keys.
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

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/batch)

# Cache key permutations

As I mentioned in the previous section, the function signature for the
`BatchFetchFn`, which the `GetOrFetchBatch` function uses, can feel too limited:

```go
type BatchFetchFn[T any] func(ctx context.Context, ids []string) (map[string]T, error)
```

What if you're fetching data from some endpoint that accepts a variety of query
parameters? Or perhaps you're doing a database query and want to apply some
ordering and filtering to the data?

Closures provide an elegant solution to this limitation. Let's illustrate this
by looking at an actual API client I've written:

```go
const moviesByIDsCacheKeyPrefix = "movies-by-ids"

type MoviesByIDsOpts struct {
	IncludeUpcoming bool
	IncludeUpsell   bool
}

func (c *Client) MoviesByIDs(ctx context.Context, ids []string, opts MoviesByIDsOpts) (map[string]Movie, error) {
	cacheKeyFunc := c.cache.PermutatedBatchKeyFn(moviesByIDsCacheKeyPrefix, opts)
	fetchFunc := func(ctx context.Context, cacheMisses []string) (map[string]Movie, error) {
		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		var response map[string]Movie
		err := requests.URL(c.baseURL).
			Path("/movies").
			Param("ids", strings.Join(cacheMisses, ",")).
			Param("include_upcoming", strconv.FormatBool(opts.IncludeUpcoming)).
			Param("include_upsell", strconv.FormatBool(opts.IncludeUpsell)).
			ToJSON(&response).
			Fetch(timeoutCtx)
		return response, err
	}
	return sturdyc.GetOrFetchBatch(ctx, c.cache, ids, cacheKeyFunc, fetchFunc)
}
```

The API clients `MoviesByIDs` method calls an external API to fetch movies by
IDs, and the `BatchFetchFn` that we're passing to `sturdyc` has a closure over
the query parameters we need.

However, one **important** thing to note here is that the ID is **no longer**
enough to _uniquely_ identify a record in our cache even with the basic prefix
function we've used before. It will no longer work to just have cache keys that
looks like this:

```
movies-ID-1
movies-ID-2
movies-ID-3
```

Now why is that? If you think about it, the query parameters will most likely
be used by the system we're calling to transform the data in various ways.
Hence, we need to store a movie not only once per ID, but also once per
transformation. In other terms, we should cache each movie once for each
permutation of our options:

```
ID 1 IncludeUpcoming: true  IncludeUpsell: true
ID 1 IncludeUpcoming: false IncludeUpsell: false
ID 1 IncludeUpcoming: true  IncludeUpsell: false
ID 1 IncludeUpcoming: false IncludeUpsell: true
```

This is what the `PermutatedBatchKeyFn` is used for. It takes a prefix and a
struct which internally it uses reflection on in order to concatenate the
**exported** fields to form a unique cache key that would look like this:

```
// movies-by-ids is our prefix that we passed as the
// first argument to the PermutatedBatchKeyFn function.
movies-by-ids-true-true-ID-1
movies-by-ids-false-false-ID-1
movies-by-ids-true-false-ID-1
movies-by-ids-false-true-ID-1
```

Please note that the struct should be flat without nesting. The fields can be
`time.Time` values, as well as any basic types, pointers to these types, and
slices containing them.

Once again, I'll provide a small example application that you can play around
with to get a deeper understanding of this functionality. We're essentially
going to use the same API client as before, but this time we're going to use
the `PermutatedBatchKeyFn` rather than the `BatchKeyFn`:

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
set. The keys would look something like this:

```
FEDEX-2024-04-06-ID-1
DHL-2024-04-07-ID-1
UPS-2024-04-08-ID-1
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

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/permutations)

# Refresh coalescing

As you may recall, our client is using the `WithEarlyRefreshes` option to
refresh the records in the background whenever their keys are requested again
after a certain amount of time has passed. And as seen in the example above,
we're successfully storing and refreshing the records once for every
permutation of the options we used to retrieve it. However, we're not taking
advantage of the endpoint's batch capabilities.

To make this more efficient, we can enable the **refresh coalescing**
functionality, but before we'll update our example to use it let's just take a
moment to understand how it works.

To start, we need to understand what determines whether two IDs can be
coalesced for a refresh: *the options*. E.g, do we want to perform the same
data transformations for both IDs? If so, they can be sent in the same batch.
This applies when we use the cache in front of a database too. Do we want to
use the same filters, sorting, etc?

If we look at the movie example from before, you can see that I've extracted
these options into a struct:

```go
const moviesByIDsCacheKeyPrefix = "movies-by-ids"

type MoviesByIDsOpts struct {
	IncludeUpcoming bool
	IncludeUpsell   bool
}

func (c *Client) MoviesByIDs(ctx context.Context, ids []string, opts MoviesByIDsOpts) (map[string]Movie, error) {
	cacheKeyFunc := c.cache.PermutatedBatchKeyFn(moviesByIDsCacheKeyPrefix, opts)
	fetchFunc := func(ctx context.Context, cacheMisses []string) (map[string]Movie, error) {
		// ...
		defer cancel()
	}
	return sturdyc.GetOrFetchBatch(ctx, c.cache, ids, cacheKeyFunc, fetchFunc)
}
```

And as I mentioned before, the `PermutatedBatchKeyFn` is going to perform
reflection on this struct to create cache keys that look something like this:

```
movies-by-ids-true-true-ID-1
movies-by-ids-false-false-ID-1
movies-by-ids-true-false-ID-1
movies-by-ids-false-true-ID-1
```

What the refresh coalescing functionality then does is that it removes the ID
but keeps the permutation string and uses it to create and uniquely
identifiable buffer where it can gather IDs that should be refreshed with the
same options:

```
movies-by-ids-true-true
movies-by-ids-false-false
movies-by-ids-true-false
movies-by-ids-false-true
```

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

So now we're saying that we want to coalesce the refreshes for each
permutation, and try to process them in batches of 3. However, if it's not able
to reach that size within 30 seconds we want the refresh to happen anyway.

The previous output revealed that the refreshes happened one by one:

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

We'll now try to run this code again, but with the `WithRefreshCoalescing`
option enabled:

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

The number of refreshes went from **9** to **3**. Imagine what a batch size of
50 could do for your applications performance!

There is more information about this in the section about metrics, but for our
production applications we're also using the `WithMetrics` option so that we
can monitor how well our refreshes are performing:

<img width="941" alt="Screenshot 2024-05-04 at 12 38 04" src="https://github.com/viccon/sturdyc/assets/12787673/b1359867-f1ef-4a09-8c75-d7d2360726f1">
This chart shows the batch sizes for our coalesced refreshes.

<img width="940" alt="Screenshot 2024-05-04 at 12 38 20" src="https://github.com/viccon/sturdyc/assets/12787673/de7f00ee-b14d-443b-b69e-91e19665c252">
This chart shows the average batch size of our refreshes for two different data sources

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/buffering)

Another point to note is how effectively the options we've seen so far can be
combined to create high-performing, flexible, and robust caching solutions:

```go
capacity := 10000
numShards := 10
ttl := 2 * time.Hour
evictionPercentage := 10
minRefreshDelay := time.Second
maxRefreshDelay := time.Second * 2
synchronousRefreshDelay := time.Second * 120 // 2 minutes.
retryBaseDelay := time.Millisecond * 10
batchSize := 10
batchBufferTimeout := time.Second * 15

cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
	sturdyc.WithEarlyRefreshes(minRefreshDelay, maxRefreshDelay, synchronousRefreshDelay, retryBaseDelay),
	sturdyc.WithRefreshCoalescing(batchSize, batchBufferTimeout),
)
```

With the configuration above, the keys in active rotation are going to be
scheduled for a refresh every 1-2 seconds. For batchable data sources, where we
are making use of the `GetOrFetchBatch` function, we'll ask the cache (using
the `WithRefreshCoalescing` option) to delay them for up to 15 seconds or until
a batch size of 10 is reached.

What if we get a request for a key that hasn't been refreshed in the last 120
seconds? Given the `synchronousRefreshDelay` passed to the `WithEarlyRefreshes`
option, the cache will skip any background refresh and instead perform a
synchronous refresh to ensure that the data is fresh. Did 1000 requests
suddenly arrive for this key? No problem, the in-flight tracking makes sure
that we only make **one** request to the underlying data source. This works for
refreshes too by the way. If 1000 requests arrived for a key that was 3 seconds
old (greater than our `maxRefreshDelay`) we would only schedule a single
refresh for it.

Is the underlying data source experiencing downtime? With our TTL of two-hours
we'll be able to provide a degraded experience to our users by serving the data
we have in our cache.

# Passthrough

There are times when you want to always retrieve the latest data from the
source and only use the in-memory cache as a _fallback_. In such scenarios, you
can use the `Passthrough` and `PassthroughBatch` functions. The cache will
still perform in-flight tracking and deduplicate your requests.

# Distributed storage

It's important to read the previous sections before jumping here in order to
understand how `sturdyc` works when it comes to creating cache keys, tracking
in-flight requests, refreshing records in the background, and
buffering/coalescing requests to minimize the number of round trips we have to
make to an underlying data source. As you'll soon see, we'll leverage all of
these features for the distributed storage too.

However, let's first understand when this functionality can be useful. This
feature is particularly valuable when building applications that can achieve a
high cache hit rate while also being subject to large bursts of requests.

As an example, I've used this in production for a large streaming application.
The content was fairly static - new movies, series, and episodes were only
ingested a couple of times an hour. This meant that we could achieve a very
high hit rate for our data sources. However, during the evenings, when a
popular football match or TV show was about to start, our traffic could spike
by a factor of 20 within less than a minute.

To illustrate the problem further, let’s say the hit rate for our in-memory
cache was 99.8%. Then, when we received that large burst of traffic, our
auto-scaling would begin provisioning new containers. These containers would
obviously be brand new, with an initial hit rate of 0%. This would cause a
significant load on our underlying data sources as soon as they came online,
because every request they received led to a cache miss so that we had to make
an outgoing request to the data source. And these data sources had gotten used
to being shielded from most of the traffic by the older containers high
hit-rate and refresh coalescing usage. Hence, what was a 20x spike for us could
easily become a 200x spike for them until our new containers had warmed their
caches.

Therefore, I decided to add the ability to have the containers sync their
in-memory cache with a distributed key-value store that would have an easier
time absorbing these bursts.

Adding distributed storage to the cache is, from the package's point of view,
essentially just another data source with a higher priority. Hence, we're still
able to take great advantage of all the features we've seen so far, and these
efficiency gains will hopefully allow us to use a much cheaper cluster.

A bit simplified, we can think of the cache's interaction with the
distributed storage like this:

```go
// NOTE: This is an example. The cache has similar functionality internally.
func (o *OrderAPI) OrderStatus(ctx context.Context, id string) (string, error) {
	cacheKey := "order-status-" + id
	fetchFn := func(ctx context.Context) (string, error) {
		// Check Redis cache first.
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

		// Add the order status to the Redis cache so that it becomes available for the other containers.
		go func() { o.RedisClient.Set(cacheKey, response.OrderStatus, time.Hour) }()

		return response.OrderStatus, nil
	}

	return o.GetOrFetch(ctx, id, fetchFn)
}
```

The real implementation interacts with the distributed storage through an
abstraction so that you're able to use any key-value store you want. All you
would have to do is implement this interface:

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
	// Other options...
	sturdyc.WithDistributedStorage(storage),
)
```

**Please note** that you are responsible for configuring the TTL and eviction
policies of this storage. `sturdyc` will only make sure that it queries this
data source first, and then writes the keys and values to this storage as soon
as it has gone out to an underlying data source and refreshed them. Therefore,
I'd advice you to use the configuration above with short TTLs for the
distributed storage, or things might get too stale. I mostly think it's useful
if you're consuming data sources that are rate limited or don't handle brief
bursts from new containers very well.

I've included an example to showcase this functionality
[here.](https://github.com/viccon/sturdyc/tree/main/examples/distribution)

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
that the remaining background refreshes that the in-memory cache performed only
went to the distributed storage.

# Distributed storage early refreshes

As I mentioned before, the configuration from the section above works well as
long as you're using short TTLs for the distributed key-value store. However,
I've also built systems where I wanted to leverage the distributed storage as
an additional robustness feature with long TTLs. That way, if an upstream
system goes down, newly provisioned containers could still retrieve the latest
data that the old containers had cached from something like a Redis.

If you have a similar use case, you could use the following
configuration instead:

```go
cacheClient := sturdyc.New[string](capacity, numShards, ttl, evictionPercentage,
	sturdyc.WithDistributedStorageEarlyRefreshes(storage, time.Minute),
)
```

With a configuration like this, I would usually set the TTL for the distributed
storage to something like an hour. However, if `sturdyc` queries the
distributed storage and finds that a record is older than 1 minute (the second
argument to the function), it will refresh the record from the underlying data
source, and then write the updated value back to it. So the interaction with
the distributed storage would look something like this:

- Start by trying to retrieve the key from the distributeted storage. If the
  data is fresh, it's returned immediately and written to the in-memory cache.
- If the key was found in the distributed storage, but wasn't fresh enough,
  we'll visit the underlying data source, and then write the response to both
  the distributed cache and the one we have in-memory.
- If the call to refresh the data failed, the cache will use the value from the
  distributed storage as a fallback.

However, there is one more scenario we must cover now that requires two
additional methods to be implemented:

```go
type DistributedStorageEarlyRefreshes interface {
	DistributedStorage
	Delete(ctx context.Context, key string)
	DeleteBatch(ctx context.Context, keys []string)
}
```

These delete methods will be called when a refresh occurs, and the cache
notices that it can no longer retrieve the key at the underlying data source.
This indicates that the key has been deleted, and we will want this change to
propagate to the distributed key-value store as soon as possible, and not have
to wait for the TTL to expire.

**Please note** that you are still responsible for setting the TTL and eviction
policies for the distributed store. The cache will only invoke the delete
methods when a record has gone missing from the underlying data source. If
you're using **missing record storage**, it will write the key as a missing
record instead.

I've included an example to showcase this functionality
[here.](https://github.com/viccon/sturdyc/tree/main/examples/distributed-early-refreshes)

# Custom metrics

The cache can be configured to report custom metrics for:

- Size of the cache
- Cache hits
- Cache misses
- Background refreshes
- Synchronous refreshes
- Missing records
- Evictions
- Forced evictions
- The number of entries evicted
- Shard distribution
- The batch size of a coalesced refresh

There are also distributed metrics if you're using the cache with a
_distributed storage_, which adds the following metrics in addition to what
we've seen above:

- Distributed cache hits
- Distributed cache misses
- Distributed refreshes
- Distributed missing records
- Distributed stale fallback

All you have to do is implement one of these interfaces:

```go
type MetricsRecorder interface {
	CacheHit()
	CacheMiss()
	AsynchronousRefresh()
	SynchronousRefresh()
	MissingRecord()
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
	DistributedRefresh()
	DistributedMissingRecord()
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

Below are a few images where some of these metrics have been visualized in Grafana:

<img width="939" alt="Screenshot 2024-05-04 at 12 36 43" src="https://github.com/viccon/sturdyc/assets/12787673/1f630aed-2322-4d3a-9510-d582e0294488">
Here we can how often we're able to serve from memory.

<img width="942" alt="Screenshot 2024-05-04 at 12 37 39" src="https://github.com/viccon/sturdyc/assets/12787673/25187529-28fb-4c4e-8fe9-9fb48772e0c0">
This image displays the number of items we have cached.

<img width="941" alt="Screenshot 2024-05-04 at 12 38 04" src="https://github.com/viccon/sturdyc/assets/12787673/b1359867-f1ef-4a09-8c75-d7d2360726f1">
This chart shows the batch sizes for the buffered refreshes.

<img width="940" alt="Screenshot 2024-05-04 at 12 38 20" src="https://github.com/viccon/sturdyc/assets/12787673/de7f00ee-b14d-443b-b69e-91e19665c252">
And lastly, we can see the average batch size of our refreshes for two different data sources.

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

- [`GetOrFetch`](https://pkg.go.dev/github.com/viccon/sturdyc#GetOrFetch)
- [`GetOrFetchBatch`](https://pkg.go.dev/github.com/viccon/sturdyc#GetOrFetchBatch)
- [`Passthrough`](https://pkg.go.dev/github.com/viccon/sturdyc#Passthrough)
- [`PassthroughBatch`](https://pkg.go.dev/github.com/viccon/sturdyc#PassthroughBatch)

They will take the cache, call the function for you, and perform the type
conversions internally. If the type conversions were to fail, you'll get a
[`ErrInvalidType`](https://pkg.go.dev/github.com/viccon/sturdyc#pkg-variables) error.

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

The entire example is available [here.](https://github.com/viccon/sturdyc/tree/main/examples/generics)
