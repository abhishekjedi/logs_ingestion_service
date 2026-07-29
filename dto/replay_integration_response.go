package dto

import "time"

type ReplayIntegrationResponse struct {
	ID                 uint64     `json:"id"`
	ProjectID          uint64     `json:"project_id"`
	Provider           string     `json:"provider"`
	ExternalProjectKey string     `json:"external_project_key"`
	APIBaseURL         string     `json:"api_base_url"`
	IngestPoint        string     `json:"ingest_point"`
	Enabled            bool       `json:"enabled"`
	HasAPIKey          bool       `json:"has_api_key"`
	LastValidatedAt    *time.Time `json:"last_validated_at,omitempty"`
	LastError          string     `json:"last_error"`
}
