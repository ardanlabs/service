package otel

import (
	"errors"

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

func endpoint(parameters trace.SamplingParameters) (string, error) {
	var path, query string

	for _, attr := range parameters.Attributes {
		switch attr.Key {
		case "url.path":
			path = attr.Value.AsString()
		case "url.query":
			query = attr.Value.AsString()
		}
	}

	if path == "" {
		return "", errors.New("url.path missing in span attribute")
	}

	if query == "" {
		return path, nil
	}

	return path + "?" + query, nil
}

// ShouldSample implements the sampler interface. It prevents the specified
// endpoints from being added to the trace.
func (ee endpointExcluder) ShouldSample(parameters trace.SamplingParameters) trace.SamplingResult {
	if ep, err := endpoint(parameters); err == nil {
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
