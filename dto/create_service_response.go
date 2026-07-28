package dto

import dbdto "error-logging/db/dto"

type CreateServiceResponse struct {
	Service   *dbdto.Service `json:"service"`
	APIKey    string         `json:"api_key"`
	IngestURL string         `json:"ingest_url"`
}
