package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

type endpointExcluder struct {
	endpoints   map[string]struct{}
	probability float64
}

func newEndpointExcluder(endpoints map[string]struct{}, probability float64) endpointExcluder {
	return endpointExcluder{
		endpoints:   endpoints,
		probability: probability,
	}
}

func endpoint(parameters trace.SamplingParameters) string {
	var path, query string

	for _, attr := range parameters.Attributes {
		switch attr.Key {
		case "url.path":
			path = attr.Value.AsString()
		case "url.query":
			query = attr.Value.AsString()
		}
	}

	switch {
	case path == "":
		return ""

	case query == "":
		return path

	default:
		return fmt.Sprintf("%s?%s", path, query)
	}
}

// ShouldSample implements the sampler interface. It prevents the specified
// endpoints from being added to the trace.
func (ee endpointExcluder) ShouldSample(parameters trace.SamplingParameters) trace.SamplingResult {
	if ep := endpoint(parameters); ep != "" {
		if _, exists := ee.endpoints[ep]; exists {
			return trace.SamplingResult{Decision: trace.Drop}
		}
	}

	return trace.TraceIDRatioBased(ee.probability).ShouldSample(parameters)
}

// Description implements the sampler interface.
func (endpointExcluder) Description() string {
	return "customSampler"
}
