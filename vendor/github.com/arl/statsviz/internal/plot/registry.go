// Package plot defines and builds the plots available in Statsviz.
package plot

import (
	"runtime/metrics"
	"slices"
	"sync"
)

type tag = string

const (
	tagGC        tag = "gc"
	tagScheduler tag = "scheduler"
	tagCPU       tag = "cpu"
	tagMisc      tag = "misc"
)

type description struct {
	metrics []string
	layout  any

	// getvalues creates the state (support struct) for the plot.
	getvalues func() getvalues
}

type registry struct {
	allMetrics   map[string]bool // names of all known runtime/metrics metrics
	metrics      []string
	descriptions []description

	samples []metrics.Sample // lazily built, only with the metrics we need
}

var reg = sync.OnceValue(func() *registry {
	reg := &registry{
		allMetrics: make(map[string]bool),
	}
	for _, m := range metrics.All() {
		reg.allMetrics[m.Name] = true
	}

	return reg
})

func (r *registry) mustidx(metric string) int {
	if !r.allMetrics[metric] {
		panic(metric + ": unknown metric in " + goversion())
	}

	idx := slices.Index(r.metrics, metric)
	if idx == -1 {
		r.metrics = append(r.metrics, metric)
		idx = len(r.metrics) - 1
	}

	return idx
}

func (r *registry) buildSamples() {
	r.samples = make([]metrics.Sample, len(r.metrics))
	for i := range r.samples {
		r.samples[i].Name = r.metrics[i]
	}
}

func (r *registry) read() []metrics.Sample {
	if r.samples == nil {
		r.buildSamples()
	}
	metrics.Read(r.samples)

	return r.samples
}

func (r *registry) register(desc description) {
	// Histograms need special handling.
	type heatmapLayoutFunc = func(samples []metrics.Sample) Heatmap
	if buildLayout, ok := desc.layout.(heatmapLayoutFunc); ok {
		// Rebuild samples to include the required metrics.
		r.buildSamples()
		samples := r.read()
		desc.layout = buildLayout(samples)
	}

	r.descriptions = append(r.descriptions, desc)

	for _, metric := range desc.metrics {
		r.mustidx(metric)
	}
}

func mustidx(metric string) int {
	// TODO: adapter for refactoring: remove
	return reg().mustidx(metric)
}

func register(desc description) struct{} {
	// TODO: adapter for refactoring: remove
	reg().register(desc)
	return struct{}{}
}
