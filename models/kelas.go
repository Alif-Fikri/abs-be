package models

import (
	"time"
)

type Kelas struct {
	ID          uint   `gorm:"primaryKey"`
	Nama        string `gorm:"type:varchar(50);not null"`
	Tingkat     string `gorm:"type:enum('SD','SMP','SMA');not null"`
	WaliKelas   *Guru  `gorm:"foreignKey:WaliKelasID"`
	WaliKelasID *uint  `gorm:"column:wali_kelas_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
