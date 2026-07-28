package dto

import (
	"time"

	"error-logging/constants"
)

type OrganizationMember struct {
	ID        uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID     uint64                 `gorm:"column:org_id" json:"org_id"`
	UserID    *uint64                `gorm:"column:user_id" json:"user_id,omitempty"`
	Email     string                 `gorm:"column:email" json:"email"`
	Role      constants.OrgRole      `gorm:"column:role" json:"role"`
	Status    constants.MemberStatus `gorm:"column:status" json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func (OrganizationMember) TableName() string { return "organization_members" }
