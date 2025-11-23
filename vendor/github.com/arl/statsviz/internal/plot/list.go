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
// and thus can't be used for a user plot.
func IsReservedPlotName(name string) bool {
	if name == "timestamp" || name == "lastgc" {
		return true
	}
	registry := reg()
	return slices.ContainsFunc(registry.descriptions, func(pd description) bool {
		return nameFromLayout(pd.layout) == name
	})
}

func nameFromLayout(layout any) string {
	switch layout := layout.(type) {
	case Scatter:
		return layout.Name
	case Heatmap:
		return layout.Name
	default:
		panic(fmt.Sprintf("unknown plot layout type %T", layout))
	}
}

// getvalues extracts, from a sample of runtime metrics, a slice with all
// the metrics necessary for a single plot.
type getvalues func(time.Time, []metrics.Sample) any

// List holds all the plots that statsviz knows about. Some plots might be
// disabled, if they rely on metrics that are unknown to the current Go version.
type List struct {
	rtPlots   []runtimePlot
	userPlots []UserPlot

	once sync.Once // ensure Config is built once
	cfg  *Config

	reg *registry
}

type runtimePlot struct {
	name    string
	getvals getvalues
	layout  any // Scatter | Heatmap
}

func NewList(userPlots []UserPlot) (*List, error) {
	if name := hasDuplicatePlotNames(userPlots); name != "" {
		return nil, fmt.Errorf("duplicate plot name %s", name)
	}

	return &List{reg: reg(), userPlots: userPlots}, nil
}

func (pl *List) enabledPlots() []runtimePlot {
	plots := make([]runtimePlot, 0, len(pl.reg.descriptions))

	for _, plot := range pl.reg.descriptions {
		var layout any
		switch l := plot.layout.(type) {
		case Scatter:
			l.Metrics = plot.metrics
			layout = l
		case Heatmap:
			l.Metrics = plot.metrics
			layout = l
		default:
			layout = plot.layout
		}

		plots = append(plots, runtimePlot{
			name:    nameFromLayout(layout),
			getvals: plot.getvalues(),
			layout:  layout,
		})
	}

	return plots
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

// WriteTo writes into w a JSON object containing the data points for all plots
// at the current instant. Return the number of written plots.
func (pl *List) WriteTo(w io.Writer) (int64, error) {
	samples := pl.reg.read()

	// lastgc time series is used as source to represent garbage collection
	// timestamps as vertical bars on certain plots.
	gcStats := debug.GCStats{}
	debug.ReadGCStats(&gcStats)

	m := map[string]any{
		// Javascript timestamps are in milliseconds.
		"lastgc": []int64{gcStats.LastGC.UnixMilli()},
	}
	now := time.Now()
	for _, p := range pl.rtPlots {
		m[p.name] = p.getvals(now, samples)
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
			Timestamp: now.UnixMilli(),
		},
	}); err != nil {
		return 0, fmt.Errorf("failed to write/convert metrics values to json: %v", err)
	}

	nplots := int64(len(pl.rtPlots) + len(pl.userPlots))
	return nplots, nil
}
