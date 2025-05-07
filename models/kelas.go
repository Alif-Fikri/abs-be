package models

import (
	"time"
)

type Kelas struct {
	ID          uint   `gorm:"primaryKey"`
	Nama        string `gorm:"type:varchar(50);not null"`
	Tingkat     string `gorm:"type:ENUM('SD','SMP','SMA');not null"`
	TahunAjaran string `gorm:"type:varchar(9);not null"`
	WaliKelasID *uint
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	WaliKelas   *Guru     `gorm:"foreignKey:WaliKelasID"`
}
