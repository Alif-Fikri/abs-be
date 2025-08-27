package models

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:255" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	Type      string    `gorm:"size:100" json:"type"`     
	Payload   string    `gorm:"type:text" json:"payload"` 
	Recipient uint      `gorm:"index" json:"recipient"`  
	Read      bool      `gorm:"default:false" json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
