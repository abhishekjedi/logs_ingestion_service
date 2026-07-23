// Package dto holds the persistence DTOs — the structs mapped onto DB tables.
// Schema itself is owned by the SQL migrations (db/migration); the gorm tags here
// only describe how these structs map onto the already-created columns.
package dto

import "time"

// Project is the top-level tenant grouping. A project owns many services.
// OwnerID is nullable for v1 until the auth/org model is decided.
type Project struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	OwnerID   *uint64   `gorm:"column:owner_id" json:"owner_id,omitempty"`
	Services  []Service `gorm:"foreignKey:ProjectID" json:"services,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Project) TableName() string { return "projects" }
