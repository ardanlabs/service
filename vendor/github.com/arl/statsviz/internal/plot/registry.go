package plot

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"runtime/metrics"
	"slices"
	"sync"
	"time"
)

// IsReservedPlotName reports whether that name is reserved for Statsviz plots
// and thus can't be chosen by user (for user plots).
func IsReservedPlotName(name string) bool {
	if name == "timestamp" || name == "lastgc" {
		return true
	}
	return slices.ContainsFunc(plotDescs, func(pd plotDesc) bool {
		return pd.name == name
	})
}

// a metricsGetter extracts, from a sample of runtime metrics, a slice with all
// the metrics necessary for a single plot.
type metricsGetter interface {
	values([]metrics.Sample) any // []uint64 | []float64
}

// List holds all the plots that statsviz knows about. Some plots might be
// disabled, if they rely on metrics that are unknown to the current Go version.
type List struct {
	rtPlots   []runtimePlot
	userPlots []UserPlot

	once sync.Once // ensure Config is built once
	cfg  *Config

	idxs        map[string]int // map metrics name to idx in samples and descs
	descs       []metrics.Description
	usedMetrics map[string]struct{}

	mu      sync.Mutex // protects samples in case of concurrent calls to WriteValues
	samples []metrics.Sample
}

type runtimePlot struct {
	name   string
	rt     metricsGetter
	layout any // Scatter | Heatmap
}

func NewList(userPlots []UserPlot) (*List, error) {
	if name := hasDuplicatePlotNames(userPlots); name != "" {
		return nil, fmt.Errorf("duplicate plot name %s", name)
	}

	descs := metrics.All()
	pl := &List{
		idxs:        make(map[string]int),
		descs:       descs,
		samples:     make([]metrics.Sample, len(descs)),
		userPlots:   userPlots,
		usedMetrics: make(map[string]struct{}),
	}
	for i := range pl.samples {
		pl.samples[i].Name = pl.descs[i].Name
	}
	metrics.Read(pl.samples)

	return pl, nil
}

func (pl *List) enabledPlots() []runtimePlot {
	plots := make([]runtimePlot, 0, len(plotDescs))

	for _, plot := range plotDescs {
		indices, enabled := pl.indicesFor(plot.metrics...)
		if enabled {
			plots = append(plots, runtimePlot{
				name:   plot.name,
				rt:     plot.make(indices...),
				layout: complete(plot.layout, plot.name, plot.tags),
			})
		}
	}

	return plots
}

// complete the layout with names and tags.
func complete(layout any, name string, tags []string) any {
	switch layout := layout.(type) {
	case Scatter:
		layout.Name = name
		layout.Tags = tags
		return layout
	case Heatmap:
		layout.Name = name
		layout.Tags = tags
		return layout
	default:
		panic(fmt.Sprintf("unknown plot layout type %T", layout))
	}
}

func (pl *List) Config() *Config {
	pl.once.Do(func() {
		pl.rtPlots = pl.enabledPlots()

		layouts := make([]any, len(pl.rtPlots))
		for i := range pl.rtPlots {
			layouts[i] = pl.rtPlots[i].layout
		}

		pl.cfg = &Config{
			Events: []string{"lastgc"},
			Series: layouts,
		}

		// User plots go at the back.
		for i := range pl.userPlots {
			pl.cfg.Series = append(pl.cfg.Series, pl.userPlots[i].Layout())
		}
	})
	return pl.cfg
}

// WriteValues writes into w a JSON object containing the data points for all
// plots at the current instant.
func (pl *List) WriteValues(w io.Writer) error {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	metrics.Read(pl.samples)

	// lastgc time series is used as source to represent garbage collection
	// timestamps as vertical bars on certain plots.
	gcStats := debug.GCStats{}
	debug.ReadGCStats(&gcStats)

	m := map[string]any{
		// Javascript timestamps are in millis.
		"lastgc": []int64{gcStats.LastGC.UnixMilli()},
	}

	for _, p := range pl.rtPlots {
		m[p.name] = p.rt.values(pl.samples)
	}

	for i := range pl.userPlots {
		up := &pl.userPlots[i]
		switch {
		case up.Scatter != nil:
			vals := make([]float64, len(up.Scatter.Funcs))
			for i := range up.Scatter.Funcs {
				vals[i] = up.Scatter.Funcs[i]()
			}
			m[up.Scatter.Plot.Name] = vals
		case up.Heatmap != nil:
			panic("unimplemented")
		}
	}

	type data struct {
		Series    map[string]any `json:"series"`
		Timestamp int64          `json:"timestamp"`
	}

	if err := json.NewEncoder(w).Encode(struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}{
		Event: "metrics",
		Data: data{
			Series:    m,
			Timestamp: time.Now().UnixMilli(),
		},
	}); err != nil {
		return fmt.Errorf("failed to write/convert metrics values to json: %v", err)
	}
	return nil
}

// indicesFor retrieves indices for the specified metrics, and a boolean
// indicating whether they were all found.
func (pl *List) indicesFor(metricNames ...string) ([]int, bool) {
	indices := make([]int, len(metricNames))
	allFound := true

	for i, name := range metricNames {
		pl.usedMetrics[name] = struct{}{} // record the metrics we use

		idx, ok := metricIdx[name]
		if !ok {
			allFound = false
		}
		indices[i] = idx
	}

	return indices, allFound
}
