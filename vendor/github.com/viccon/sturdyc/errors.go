package sturdyc

import "errors"

var (
	// errOnlyDistributedRecords is an internal error that the cache uses to not
	// store records as missing if it's unable to get part of the batch from the
	// underlying data source.
	errOnlyDistributedRecords = errors.New("sturdyc: we were only able to get records from the distributed storage")
	// ErrNotFound should be returned from a FetchFn to indicate that a record is
	// missing at the underlying data source. This helps the cache to determine
	// if a record should be deleted or stored as a missing record if you have
	// that functionality enabled. Missing records are refreshed like any other
	// record, and if your FetchFn returns a value for it, the record will no
	// longer be considered missing. Please note that this only applies to
	// client.GetOrFetch and client.Passthrough. For client.GetOrFetchBatch and
	// client.PassthroughBatch, this works implicitly if you return
	// a map without the ID, and have store missing records enabled.
	ErrNotFound = errors.New("sturdyc: err not found")
	// ErrMissingRecord is returned by client.GetOrFetch and client.Passthrough when a record has been marked
	// as missing. The cache will still try to refresh the record in the background if it's being requested.
	ErrMissingRecord = errors.New("sturdyc: the record has been marked as missing in the cache")
	// ErrOnlyCachedRecords is returned by client.GetOrFetchBatch and
	// client.PassthroughBatch when some of the requested records are available
	// in the cache, but the attempt to fetch the remaining records failed. It
	// may also be returned when you're using the WithEarlyRefreshes
	// functionality, and the call to synchronously refresh a record failed. The
	// cache will then give you the latest data it has cached, and you as the
	// consumer can then decide whether to proceed with the cached records or if
	// the newest data is necessary.
	ErrOnlyCachedRecords = errors.New("sturdyc: failed to fetch the records that were not in the cache")
	// ErrInvalidType is returned when you try to use one of the generic
	// package level functions but the type assertion fails.
	ErrInvalidType = errors.New("sturdyc: invalid response type")
)
