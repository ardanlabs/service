package sturdyc

import (
	"time"
)

// buffer represents a buffer for a batch refresh.
type buffer struct {
	channel chan []string
	ids     []string
}

// createBuffer should be called WITH a lock when a refresh buffer is created.
func (c *Client[T]) createBuffer(permutation string, ids []string) {
	bufferIDs := make([]string, 0, c.bufferSize)
	bufferIDs = append(bufferIDs, ids...)
	buf := &buffer{
		channel: make(chan []string),
		ids:     bufferIDs,
	}
	c.permutationBufferMap[permutation] = buf
}

// deleteBuffer should be called WITH a lock when a buffer has been processed.
func (c *Client[T]) deleteBuffer(permutation string) {
	delete(c.permutationBufferMap, permutation)
}

// bufferBatchRefresh will buffer the batch of IDs until the batch size is reached or the buffer duration is exceeded.
func bufferBatchRefresh[T any](c *Client[T], ids []string, keyFn KeyFn, fetchFn BatchFetchFn[T]) {
	if len(ids) == 0 {
		return
	}

	// If we got a perfect batch size, we can refresh the records immediately.
	if len(ids) == c.bufferSize {
		c.refreshBatch(ids, keyFn, fetchFn)
		return
	}

	c.batchMutex.Lock()

	// If the ids are greater than our batch size we'll have to chunk them.
	if len(ids) > c.bufferSize {
		idsToRefresh := ids[:c.bufferSize]
		overflowingIDs := ids[c.bufferSize:]
		c.batchMutex.Unlock()

		// These IDs are the size we want, so we'll refresh them immediately.
		c.safeGo(func() {
			c.refreshBatch(idsToRefresh, keyFn, fetchFn)
		})

		// We'll continue to process the remaining IDs recursively.
		c.safeGo(func() {
			bufferBatchRefresh(c, overflowingIDs, keyFn, fetchFn)
		})

		return
	}

	// Extract the permutation string from the ids.
	permutationString := extractPermutation(keyFn(ids[0]))

	// Check if we already have a batch for this set of options.
	if buf, ok := c.permutationBufferMap[permutationString]; ok {
		// There is a small chance that another goroutine manages to write to the channel
		// and fill the buffer as we unlock this mutex. Therefore, we'll add a timer so
		// that we can process these ids again if that were to happen.
		c.batchMutex.Unlock()
		timer, stop := c.clock.NewTimer(time.Millisecond * 10)
		select {
		case buf.channel <- ids:
			stop()
		case <-timer:
			c.safeGo(func() {
				bufferBatchRefresh(c, ids, keyFn, fetchFn)
			})
		}
		return
	}

	// There is no existing batch buffering for this permutation
	// of options. Hence, we'll create a new one.
	c.createBuffer(permutationString, ids)
	c.batchMutex.Unlock()

	c.safeGo(func() {
		timer, stop := c.clock.NewTimer(c.bufferTimeout)
		c.batchMutex.Lock()
		idStream := c.permutationBufferMap[permutationString].channel
		c.batchMutex.Unlock()

		for {
			select {
			// If the buffer times out, we'll refresh the records regardless of the buffer size.
			case _, ok := <-timer:
				if !ok {
					return
				}

				// We reached the deadline for this batch.
				c.batchMutex.Lock()
				buffer := c.permutationBufferMap[permutationString]
				c.deleteBuffer(permutationString)
				c.batchMutex.Unlock()

				c.safeGo(func() {
					c.refreshBatch(buffer.ids, keyFn, fetchFn)
				})
				return

			case additionalIDs, ok := <-idStream:
				if !ok {
					return
				}

				// Lock the mutex, and add the additional IDs to the buffer.
				c.batchMutex.Lock()
				buffer := c.permutationBufferMap[permutationString]
				buffer.ids = append(buffer.ids, additionalIDs...)

				// If we haven't reached the batch size yet, we'll wait for more ids.
				if len(buffer.ids) < c.bufferSize {
					c.batchMutex.Unlock()
					continue
				}

				// At this point, we have either reached or exceeded the batch size. We'll stop the timer and drain the channel.
				if !stop() {
					<-timer
				}

				// Grab a reference to the IDs, and then delete the buffer.
				permIDs := buffer.ids
				c.deleteBuffer(permutationString)
				c.batchMutex.Unlock()

				idsToRefresh := permIDs[:c.bufferSize]
				overflowingIDs := permIDs[c.bufferSize:]

				// Refresh the first batch of IDs immediately.
				c.safeGo(func() {
					c.refreshBatch(idsToRefresh, keyFn, fetchFn)
				})

				// If we exceeded the batch size, we'll continue to process the remaining IDs recursively.
				if len(overflowingIDs) > 0 {
					c.safeGo(func() {
						bufferBatchRefresh(c, overflowingIDs, keyFn, fetchFn)
					})
				}
				return
			}
		}
	})
}
