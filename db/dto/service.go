package dto

import "time"

// Service is a single application/service registered under a project. Logs are
// pushed to a per-service ingest URL keyed by PublicID and authenticated with an
// API key. Only the sha256 hash of the key is stored; the raw key is shown once at
// creation. APIKeyPrefix is a non-secret fragment kept for display.
type Service struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID    uint64    `gorm:"column:project_id" json:"project_id"`
	Name         string    `gorm:"column:name" json:"name"`
	PublicID     string    `gorm:"column:public_id" json:"public_id"`
	APIKeyHash   string    `gorm:"column:api_key_hash" json:"-"`
	APIKeyPrefix string    `gorm:"column:api_key_prefix" json:"api_key_prefix"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Service) TableName() string { return "services" }
