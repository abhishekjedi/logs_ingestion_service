package dto

import "time"

// Organization is the top-level tenant. Users are members of orgs; projects belong
// to an org.
type Organization struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Slug      string    `gorm:"column:slug" json:"slug"`
	CreatedBy uint64    `gorm:"column:created_by" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Organization) TableName() string { return "organizations" }
