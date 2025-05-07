package models

import (
	"time"
)

type Guru struct {
	ID           uint       `gorm:"primaryKey"`
	Nama         string     `gorm:"type:varchar(100);not null"`
	NIP          string     `gorm:"type:varchar(30);unique;not null"`
	NIK          string     `gorm:"type:varchar(30);unique;not null"`
	Email        string     `gorm:"type:varchar(100);unique;not null"`
	Telepon      string     `gorm:"type:varchar(20)"`
	Alamat       string     `gorm:"type:text"`
	JenisKelamin string     `gorm:"type:enum('L','P');not null"`
	Password     string     `gorm:"type:varchar(255);not null"`
	GuruRoles    []GuruRole `gorm:"foreignKey:GuruID"`
	KelasWali    []Kelas    `gorm:"foreignKey:WaliKelasID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GuruRole struct {
	ID        uint   `gorm:"primaryKey"`
	GuruID    uint   `gorm:"not null"`
	Role      string `gorm:"type:ENUM('wali_kelas','guru_mapel');not null"`
	KelasID   *uint
	MapelID   *uint
	CreatedAt time.Time     `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time     `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	Guru      Guru          `gorm:"foreignKey:GuruID"`
	Kelas     Kelas         `gorm:"foreignKey:KelasID"`
	Mapel     MataPelajaran `gorm:"foreignKey:MapelID"`
}

type GuruMapelKelas struct {
	ID            uint `gorm:"primaryKey"`
	GuruID        uint `gorm:"not null;uniqueIndex:idx_guru_mapel_kelas"`
	Guru          Guru
	MapelID       uint `gorm:"not null;uniqueIndex:idx_guru_mapel_kelas"`
	MataPelajaran MataPelajaran
	KelasID       uint `gorm:"not null;uniqueIndex:idx_guru_mapel_kelas"`
	Kelas         Kelas
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
