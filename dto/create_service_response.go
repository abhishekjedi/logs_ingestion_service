package dto

import dbdto "error-logging/db/dto"

// CreateServiceResponse returns the freshly minted service plus the raw API key.
// The raw key is shown exactly once here and never persisted or returned again.
type CreateServiceResponse struct {
	Service   *dbdto.Service `json:"service"`
	APIKey    string         `json:"api_key"`
	IngestURL string         `json:"ingest_url"`
}
