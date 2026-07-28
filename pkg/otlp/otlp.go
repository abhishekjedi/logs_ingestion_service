package otlp

import "strconv"

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

type AnyValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    *string  `json:"intValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
}

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
