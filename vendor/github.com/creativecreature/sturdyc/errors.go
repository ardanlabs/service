package sturdyc

import "errors"

var (
	// ErrStoreMissingRecord should be returned from a FetchFn to indicate that a record is
	// to be marked as missing by the cache. This will prevent continuous outgoing requests
	// to the source. The record will still be refreshed like any other record, and if your
	// FetchFn returns a value for it, the record will no longer be considered missing.
	// Please note that this only applies to client.GetFetch and client.Passthrough. For
	// client.GetFetchBatch and client.PassthroughBatch, this works implicitly if you return
	// a map without the ID, and have store missing records enabled.
	ErrStoreMissingRecord = errors.New("the record will be marked as missing in the cache")
	// ErrMissingRecord is returned by client.GetFetch and client.Passthrough when a record has been marked
	// as missing. The cache will still try to refresh the record in the background if it's being requested.
	ErrMissingRecord = errors.New("the record has been marked as missing in the cache")
	// ErrOnlyCachedRecords is returned by client.GetFetchBatch and client.PassthroughBatch
	// when some of the requested records are available in the cache, but the attempt to
	// fetch the remaining records failed. As the consumer, you can then decide whether to
	// proceed with the cached records or if the entire batch is necessary.
	ErrOnlyCachedRecords = errors.New("failed to fetch the records that were not in the cache")
	// ErrInvalidType is returned when you try to use one of the generic
	// package level functions but the type assertion fails.
	ErrInvalidType = errors.New("invalid response type")
	// ErrDeletedRecord should be returned by the FetchFn that is passed to client.GetFetch
	// and client.Passthrough if you wan't to have a record deleted by a refresh. This is
	// not needed for client.GetFetchBatch and client.PassthroughBatch as the cache will
	// delete the record automatically if stampede protection is enabled and store missing
	// records is set to false.
	ErrDeleteRecord = errors.New("the record has been deleted at the data source")
)
