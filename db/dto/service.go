package dto

import "time"

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
