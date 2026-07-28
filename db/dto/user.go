package dto

import "time"

type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"column:email" json:"email"`
	Name      string    `gorm:"column:name" json:"name"`
	AvatarURL string    `gorm:"column:avatar_url" json:"avatar_url"`
	GoogleSub *string   `gorm:"column:google_sub" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }
