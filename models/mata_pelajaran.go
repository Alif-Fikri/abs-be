package models

import (
	"time"
)

type MataPelajaran struct {
	ID           uint   `gorm:"primaryKey"`
	Nama         string `gorm:"type:varchar(100);not null"`
	Kode         string `gorm:"type:varchar(20);unique;not null"`
	Tingkat      string `gorm:"type:enum('SD','SMP','SMA');default:'SMP'not null"`
	GuruPengampu *Guru  `gorm:"foreignKey:GuruID"`
	GuruID       *uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
