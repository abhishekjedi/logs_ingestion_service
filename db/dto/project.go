package dto

import "time"

type Project struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID     *uint64   `gorm:"column:org_id" json:"org_id,omitempty"`
	Name      string    `gorm:"column:name" json:"name"`
	OwnerID   *uint64   `gorm:"column:owner_id" json:"owner_id,omitempty"`
	Services  []Service `gorm:"foreignKey:ProjectID" json:"services,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Project) TableName() string { return "projects" }
