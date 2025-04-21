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
	// ErrOnlyCachedRecords can be returned when you're using the cache with
	// early refreshes or distributed storage functionality. It indicates that
	// the records *should* have been refreshed from the underlying data source,
	// but the operation failed. It is up to you to decide whether you want to
	// proceed with the records that were retrieved from the cache. Note: For
	// batch operations, this might contain only part of the batch. For example,
	// if you requested keys 1-10, and we had IDs 1-3 in the cache, but the
	// request to fetch records 4-10 failed.
	ErrOnlyCachedRecords = errors.New("sturdyc: failed to fetch the records that were not in the cache")
	// ErrInvalidType is returned when you try to use one of the generic
	// package level functions but the type assertion fails.
	ErrInvalidType = errors.New("sturdyc: invalid response type")
)

// onlyCachedRecords is used when we were able to successfully retrieve some
// records from distributed storage, but the request to get additional records
// from the underlying data source failed. In this case, we wrap any potential
// errors from the underlying data source with an ErrOnlyCachedRecords to allow
// the user to decide whether to proceed with the cached records or not.
func onlyCachedRecords(err error) error {
	multiErr, isMultiErr := err.(interface{ Unwrap() []error })
	if !isMultiErr {
		if errors.Is(errOnlyDistributedRecords, err) {
			return ErrOnlyCachedRecords
		}
		return errors.Join(err, ErrOnlyCachedRecords)
	}

	var errs []error
	errs = append(errs, ErrOnlyCachedRecords)
	for _, e := range multiErr.Unwrap() {
		if !errors.Is(e, errOnlyDistributedRecords) {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}
