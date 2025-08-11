package models

import (
	"time"
)

type MataPelajaran struct {
	ID         uint      `gorm:"primaryKey"`
	Nama       string    `gorm:"type:varchar(100);not null"`
	Kode       string    `gorm:"type:varchar(20);unique;not null"`
	Tingkat    string    `gorm:"type:enum('SD','SMP','SMA');default:'SMP';not null"`
	JamMulai   time.Time `gorm:"type:time" json:"jam_mulai"`
	JamSelesai time.Time `gorm:"type:time" json:"jam_selesai"`
	Hari       string    `gorm:"type:enum('Senin','Selasa','Rabu','Kamis','Jumat','Sabtu');default:'Senin'" json:"hari"`
	Semester   string    `gorm:"type:enum('ganjil','genap')" json:"semester"`
	IsActive   bool      `gorm:"type:boolean;default:true" json:"is_active"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
