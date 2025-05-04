package models

import (
	"time"
)

type Admin struct {
	ID        uint   `gorm:"primaryKey"`
	Nama      string `gorm:"type:varchar(100);not null"`
	Email     string `gorm:"type:varchar(100);unique;not null"`
	Password  string `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
