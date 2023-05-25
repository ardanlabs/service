package debug

import (
	"expvar"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type c struct{}

func (c c) Collect(ch chan<- prometheus.Metric) {
	values := map[string]interface{}{}

	expvar.Do(func(kv expvar.KeyValue) {
		values[fmt.Sprintf("expvar_%s", kv.Key)] = kv.Value
	})

	c.do(ch, values)
}

func (c c) do(ch chan<- prometheus.Metric, values map[string]interface{}) {
	for k, v := range values {
		name := strings.ReplaceAll(k, "/", "_")
		desc := prometheus.NewDesc(name, "", nil, nil)

		switch v := v.(type) {
		case *expvar.Int:
			ch <- prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, toFloat64(v.Value()))
		case *expvar.Float:
			ch <- prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, v.Value())
		case *expvar.Map:
			values := map[string]interface{}{}

			v.Do(func(kv expvar.KeyValue) {
				values[fmt.Sprintf("%s_%s", k, kv.Key)] = kv.Value
			})
			c.do(ch, values)
		default:
			// TODO: support expvar.String, expvar.Func
			continue
		}
	}
}

func toFloat64(v interface{}) float64 {
	switch v := v.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	}
	panic(fmt.Sprintf("unexpected value type: %#v", v))
}

func (c c) Describe(chan<- *prometheus.Desc) {}
