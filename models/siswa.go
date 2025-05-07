package models

import (
	"time"
)

type Siswa struct {
	ID           uint      `gorm:"primaryKey"`
	Nama         string    `gorm:"type:varchar(100);not null"`
	NISN         string    `gorm:"type:varchar(20);unique;not null"`
	TempatLahir  string    `gorm:"type:varchar(100);not null"`
	TanggalLahir time.Time `gorm:"type:date"`
	JenisKelamin string    `gorm:"type:enum('L','P');not null"`
	NamaAyah     string    `gorm:"type:varchar(100);not null"`
	NamaIbu      string    `gorm:"type:varchar(100);not null"`
	Alamat       string    `gorm:"type:text;not null"`
	Agama        string    `gorm:"type:varchar(20);not null"`
	Email        string    `gorm:"type:varchar(100);unique"`
	Telepon      string    `gorm:"type:varchar(20)"`
	AsalSekolah  string    `gorm:"type:varchar(100);not null"`
	KelasID      uint
	Kelas        Kelas `gorm:"foreignKey:KelasID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
