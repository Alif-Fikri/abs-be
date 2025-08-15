package models

import (
	"time"
)

type Siswa struct {
	ID            uint             `gorm:"primaryKey"`
	Nama          string           `gorm:"type:varchar(100);not null"`
	NISN          string           `gorm:"type:varchar(20);unique;not null"`
	Password      string           `gorm:"type:varchar(255);not null"`
	TempatLahir   string           `gorm:"type:varchar(100);not null"`
	TanggalLahir  time.Time        `gorm:"type:date"`
	JenisKelamin  string           `gorm:"type:enum('L','P');not null"`
	NamaAyah      string           `gorm:"type:varchar(100);not null"`
	NamaIbu       string           `gorm:"type:varchar(100);not null"`
	Alamat        string           `gorm:"type:text;not null"`
	Agama         string           `gorm:"type:varchar(20);not null"`
	Email         string           `gorm:"type:varchar(100);unique"`
	Telepon       string           `gorm:"type:varchar(20)"`
	AsalSekolah   string           `gorm:"type:varchar(100);not null"`
	KelasID       *uint            `gorm:"index"`
	Kelas         []*Kelas         `gorm:"many2many:kelas_siswas;"`
	MataPelajaran []*MataPelajaran `gorm:"many2many:mapel_siswas;joinForeignKey:SiswaID;joinReferences:MapelID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
