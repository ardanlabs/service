package sturdyc

import "time"

// deleteBuffer should be called WITH a lock when a buffer has been processed.
func deleteBuffer[T any](c *Client[T], batchIdentifier string) {
	delete(c.bufferPermutationChan, batchIdentifier)
	delete(c.bufferPermutationIDs, batchIdentifier)
}

// bufferBatchRefresh will buffer the batch of IDs until the batch size is reached or the buffer duration is exceeded.
func bufferBatchRefresh[T any](c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) {
	if len(ids) == 0 {
		return
	}

	// If we got a perfect batch size, we can refresh the records immediately.
	if len(ids) == c.batchSize {
		c.refreshBatch(ids, keyFn, fetchFn)
		return
	}

	c.batchMutex.Lock()

	// If the ids are greater than our batch size we'll have to chunk them.
	if len(ids) > c.batchSize {
		idsToRefresh := ids[:c.batchSize]
		overflowingIDs := ids[c.batchSize:]
		c.batchMutex.Unlock()

		// These IDs are the size we want, so we'll refresh them immediately.
		safeGo(func() {
			c.refreshBatch(idsToRefresh, keyFn, fetchFn)
		})

		// We'll continue to process the remaining IDs recursively.
		safeGo(func() {
			bufferBatchRefresh(c, overflowingIDs, keyFn, fetchFn)
		})

		return
	}

	// Extract the permutation string from the ids.
	permutationString := extractPermutation(keyFn(ids[0]))

	// Check if we already have a batch for this set of options.
	if channel, ok := c.bufferPermutationChan[permutationString]; ok {
		// There is a small chance that another goroutine manages to write to the channel
		// and fill the buffer as we unlock this mutex. Therefore, we'll add a timer so
		// that we can process these ids again if that were to happen.
		c.batchMutex.Unlock()
		timer, stop := c.clock.NewTimer(time.Millisecond * 10)
		select {
		case channel <- ids:
			stop()
		case <-timer:
			safeGo(func() {
				bufferBatchRefresh(c, ids, keyFn, fetchFn)
			})
			return
		}
		return
	}

	// There is no existing batch buffering for this permutation
	// of options. Hence, we'll create a new one.
	idStream := make(chan []string)
	c.bufferPermutationChan[permutationString] = idStream
	c.bufferPermutationIDs[permutationString] = ids
	c.batchMutex.Unlock()

	safeGo(func() {
		timer, stop := c.clock.NewTimer(c.bufferTimeout)

		for {
			select {
			// If the buffer times out, we'll refresh the records regardless of the buffer size.
			case _, ok := <-timer:
				if !ok {
					return
				}

				// We reached the deadline for this batch.
				c.batchMutex.Lock()
				permIDs := c.bufferPermutationIDs[permutationString]
				deleteBuffer(c, permutationString)
				c.batchMutex.Unlock()

				safeGo(func() {
					c.refreshBatch(permIDs, keyFn, fetchFn)
				})
				return

			case additionalIDs, ok := <-idStream:
				if !ok {
					return
				}

				// Lock the mutex, and add the additional IDs to the buffer.
				c.batchMutex.Lock()
				c.bufferPermutationIDs[permutationString] = append(c.bufferPermutationIDs[permutationString], additionalIDs...)

				// If we haven't reached the batch size yet, we'll wait for more ids.
				if len(c.bufferPermutationIDs[permutationString]) < c.batchSize {
					c.batchMutex.Unlock()
					continue
				}

				// At this point, we have either reached or exceeded the batch size. We'll stop the timer and drain the channel.
				if !stop() {
					<-timer
				}

				// Grab a reference to the IDs, and then delete the buffer.
				permIDs := c.bufferPermutationIDs[permutationString]
				deleteBuffer(c, permutationString)
				c.batchMutex.Unlock()

				idsToRefresh := permIDs[:c.batchSize]
				overflowingIDs := permIDs[c.batchSize:]

				// Refresh the first batch of IDs immediately.
				safeGo(func() {
					c.refreshBatch(idsToRefresh, keyFn, fetchFn)
				})

				// If we exceeded the batch size, we'll continue to process the remaining IDs recursively.
				if len(overflowingIDs) > 0 {
					safeGo(func() {
						bufferBatchRefresh(c, overflowingIDs, keyFn, fetchFn)
					})
				}
				return
			}
		}
	})
}
