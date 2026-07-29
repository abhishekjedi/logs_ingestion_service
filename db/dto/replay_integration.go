package dto

import "time"

type ReplayIntegration struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID          uint64     `gorm:"column:project_id" json:"project_id"`
	Provider           string     `gorm:"column:provider" json:"provider"`
	ExternalProjectKey string     `gorm:"column:external_project_key" json:"external_project_key"`
	APIBaseURL         string     `gorm:"column:api_base_url" json:"api_base_url"`
	IngestPoint        string     `gorm:"column:ingest_point" json:"ingest_point"`
	APIKeyCiphertext   string     `gorm:"column:api_key_ciphertext" json:"-"`
	Enabled            bool       `gorm:"column:enabled" json:"enabled"`
	LastValidatedAt    *time.Time `gorm:"column:last_validated_at" json:"last_validated_at,omitempty"`
	LastError          string     `gorm:"column:last_error" json:"last_error"`
	CreatedAt          time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (ReplayIntegration) TableName() string { return "replay_integrations" }
