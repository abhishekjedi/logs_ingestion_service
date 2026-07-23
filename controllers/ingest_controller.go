package controllers

import "error-logging/pkg/context"

type IngestController interface {
	Ingest(c *context.ApiContext)
}
