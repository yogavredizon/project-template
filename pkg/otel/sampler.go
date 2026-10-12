package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
)

type exluderEndpoints struct {
	endpoints   map[string]struct{}
	probability float64
}

func newExcluderEndpoints(endpoints map[string]struct{}, probability float64) exluderEndpoints {
	return exluderEndpoints{
		endpoints:   endpoints,
		probability: probability,
	}
}

func endpoints(params trace.SamplingParameters) string {
	var path, query string
	for _, attr := range params.Attributes {
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
		return ""
	default:
		return fmt.Sprintf("%s?%s", path, query)
	}
}

func (e exluderEndpoints) ShouldSample(params trace.SamplingParameters) trace.SamplingResult {
	if ep := endpoints(params); ep != "" {
		if _, ok := e.endpoints[ep]; ok {
			return trace.SamplingResult{Decision: trace.Drop}
		}
	}

	return trace.TraceIDRatioBased(e.probability).ShouldSample(params)
}

func (e exluderEndpoints) Description() string {
	return "customSample"
}
