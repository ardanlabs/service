package sturdyc

import "errors"

var (
	// ErrStoreMissingRecord should be returned from FetchFn to indicate that we
	// want to store the record with a cooldown. This only applies to the
	// FetchFn, for the BatchFetchFn you should enable the functionality through
	// options, and simply return a map without the missing record being present.
	ErrStoreMissingRecord = errors.New("record not found")
	// ErrMissingRecord is returned by sturdyc.GetFetch when a record has been fetched unsuccessfully.
	ErrMissingRecord = errors.New("record is missing")
	// ErrOnlyCachedRecords is returned by sturdyc.GetFetchBatch when we have
	// some of the requested records in the cache, but the call to fetch the
	// remaining records failed. The consumer can then choose if they want to
	// proceed with the cached records or retry the operation.
	ErrOnlyCachedRecords = errors.New("failed to fetch the records that we did not have cached")
	// ErrInvalidType is returned when you try to use one of the generic
	// package level functions, and the type assertion fails.
	ErrInvalidType = errors.New("invalid response type")
)

func ErrIsStoreMissingRecordOrMissingRecord(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrStoreMissingRecord) || errors.Is(err, ErrMissingRecord)
}
