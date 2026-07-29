package dto

type UpsertReplayIntegrationRequest struct {
	ExternalProjectKey string `json:"external_project_key" binding:"required"`
	APIBaseURL         string `json:"api_base_url"`
	IngestPoint        string `json:"ingest_point"`
	OrganizationAPIKey string `json:"organization_api_key"`
	Enabled            *bool  `json:"enabled"`
}
