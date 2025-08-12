package models

import (
	"time"
)

type Guru struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Nama         string     `gorm:"type:varchar(100);not null" json:"nama"`
	NIP          string     `gorm:"column:nip;type:varchar(30);unique;not null" json:"nip"`
	NIK          string     `gorm:"column:nik;type:varchar(30);unique;not null" json:"nik"`
	Email        string     `gorm:"type:varchar(100);unique;not null" json:"email"`
	Telepon      string     `gorm:"type:varchar(20)" json:"telepon,omitempty"`
	Alamat       string     `gorm:"type:text" json:"alamat,omitempty"`
	JenisKelamin string     `gorm:"type:enum('L','P');not null" json:"jenis_kelamin"`
	Password     string     `gorm:"type:varchar(255);not null" json:"-"`
	GuruRoles    []GuruRole `gorm:"foreignKey:GuruID" json:"guru_roles,omitempty"`
	KelasWali    []Kelas    `gorm:"foreignKey:WaliKelasID" json:"kelas_wali,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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
	TahunAjaran   string `gorm:"type:varchar(9);not null;uniqueIndex:idx_guru_mapel_kelas"`
	Semester      string `gorm:"type:enum('ganjil','genap');not null;uniqueIndex:idx_guru_mapel_kelas"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
