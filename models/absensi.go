package models

import (
	"time"
)

type AbsensiSiswa struct {
	ID            uint `gorm:"primaryKey"`
	SiswaID       uint `gorm:"not null"`
	Siswa         Siswa
	KelasID       uint `gorm:"not null"`
	Kelas         Kelas
	MapelID       *uint // bisa null jika absensi kelas
	MataPelajaran *MataPelajaran
	GuruID        uint `gorm:"not null"`
	Guru          Guru
	TipeAbsensi   string    `gorm:"type:enum('kelas','mapel');not null"`
	Tanggal       time.Time `gorm:"type:date;not null"`
	Status        string    `gorm:"type:enum('masuk','izin','sakit','terlambat','alpa');not null"`
	Keterangan    string    `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
