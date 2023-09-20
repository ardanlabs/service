// Copyright 2019, LightStep Inc.

package internal

type Vsample struct {
	Sample interface{}
	Weight float64
}

type SampleHeap []Vsample

func (sh *SampleHeap) Push(v Vsample) {
	l := append(*sh, v)
	n := len(l) - 1

	// This copies the body of heap.up().
	j := n
	for {
		i := (j - 1) / 2 // parent
		if i == j || l[j].Weight >= l[i].Weight {
			break
		}
		l[i], l[j] = l[j], l[i]
		j = i
	}

	*sh = l
}

func (sh *SampleHeap) Pop() Vsample {
	l := *sh
	n := len(l) - 1
	result := l[0]
	l[0] = l[n]
	l = l[:n]

	// This copies the body of heap.down().
	i := 0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && l[j2].Weight < l[j1].Weight {
			j = j2 // = 2*i + 2  // right child
		}
		if l[j].Weight >= l[i].Weight {
			break
		}
		l[i], l[j] = l[j], l[i]
		i = j
	}

	*sh = l
	return result
}
