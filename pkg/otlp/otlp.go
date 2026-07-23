// Package otlp models the subset of the OTLP/JSON logs wire format we consume and
// flattens it into normalized records for the pipeline. It intentionally hand-rolls
// the structs (rather than pulling the full proto module) since we only read a
// handful of fields; the wire format we accept is still standard OTLP/JSON.
package otlp

import "strconv"

// ExportLogsRequest is the top-level OTLP logs payload (ExportLogsServiceRequest).
type ExportLogsRequest struct {
	ResourceLogs []ResourceLogs `json:"resourceLogs"`
}

type ResourceLogs struct {
	Resource  Resource    `json:"resource"`
	ScopeLogs []ScopeLogs `json:"scopeLogs"`
}

type Resource struct {
	Attributes []KeyValue `json:"attributes"`
}

type ScopeLogs struct {
	LogRecords []LogRecord `json:"logRecords"`
}

type LogRecord struct {
	TimeUnixNano         string     `json:"timeUnixNano"`
	ObservedTimeUnixNano string     `json:"observedTimeUnixNano"`
	SeverityNumber       int        `json:"severityNumber"`
	SeverityText         string     `json:"severityText"`
	Body                 AnyValue   `json:"body"`
	Attributes           []KeyValue `json:"attributes"`
	TraceID              string     `json:"traceId"`
	SpanID               string     `json:"spanId"`
}

type KeyValue struct {
	Key   string   `json:"key"`
	Value AnyValue `json:"value"`
}

// AnyValue is the OTLP union type. We support the scalar variants; complex ones
// (array/kvlist) stringify to empty, which is fine for the attributes we extract.
type AnyValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    *string  `json:"intValue,omitempty"` // int64 encoded as string
	BoolValue   *bool    `json:"boolValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
}

// AsString renders the scalar value as a string.
func (v AnyValue) AsString() string {
	switch {
	case v.StringValue != nil:
		return *v.StringValue
	case v.IntValue != nil:
		return *v.IntValue
	case v.BoolValue != nil:
		return strconv.FormatBool(*v.BoolValue)
	case v.DoubleValue != nil:
		return strconv.FormatFloat(*v.DoubleValue, 'g', -1, 64)
	default:
		return ""
	}
}

// asMap collapses a KeyValue list into a plain string map.
func asMap(kvs []KeyValue) map[string]string {
	if len(kvs) == 0 {
		return map[string]string{}
	}
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Value.AsString()
	}
	return m
}
