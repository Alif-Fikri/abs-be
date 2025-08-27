package models

import "time"

type DeviceToken struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserID    uint   `gorm:"index" json:"user_id"` 
	Token     string `gorm:"size:512;uniqueIndex:idx_user_token" json:"token"`
	Platform  string `gorm:"size:50" json:"platform,omitempty"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
