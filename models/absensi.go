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
	Semester      string    `gorm:"type:enum('ganjil','genap');not null"`
	TahunAjaran   string    `gorm:"type:varchar(9);not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AbsensiResult struct {
	ID          uint      `json:"id"`
	SiswaID     uint      `json:"siswa_id"`
	NamaSiswa   string    `json:"nama_siswa"`
	KelasID     uint      `json:"kelas_id"`
	MapelID     *uint     `json:"mapel_id,omitempty"`
	NamaMapel   *string   `json:"nama_mapel,omitempty"`
	GuruID      uint      `json:"guru_id"`
	TipeAbsensi string    `json:"tipe_absensi"`
	Tanggal     string    `json:"tanggal"`
	Status      string    `json:"status"`
	Keterangan  string    `json:"keterangan,omitempty"`
	TahunAjaran string    `json:"tahun_ajaran"`
	Semester    string    `json:"semester"`
}
