package sturdyc

import "time"

func partition(times []time.Time, low, high int) int {
	pivot := times[high]
	i := low
	for j := low; j < high; j++ {
		if times[j].Before(pivot) {
			times[i], times[j] = times[j], times[i]
			i++
		}
	}
	times[i], times[high] = times[high], times[i]
	return i
}

func quickSelect(times []time.Time, low, high, k int) time.Time {
	if low < high {
		pi := partition(times, low, high)
		if pi == k {
			return times[pi]
		} else if pi < k {
			return quickSelect(times, pi+1, high, k)
		}
		return quickSelect(times, low, pi-1, k)
	}

	// Base case for single element
	return times[low]
}

// FindCutoff returns the time that is the k-th smallest time in the slice.
func FindCutoff(times []time.Time, percentile float64) time.Time {
	if len(times) == 0 {
		return time.Time{}
	}
	if percentile < 0 || percentile > 1 {
		return time.Time{}
	}

	n := len(times)
	// Calculate the index for the given percentile
	k := int(float64(n) * percentile)
	// Adjust if k equals the length of the slice
	if k == n {
		k--
	}
	return quickSelect(times, 0, n-1, k)
}
