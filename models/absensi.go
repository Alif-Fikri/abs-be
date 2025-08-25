package models

import (
	"time"
)

type AbsensiSiswa struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	SiswaID       uint           `gorm:"not null;column:siswa_id" json:"siswa_id"`
	Siswa         Siswa          `gorm:"foreignKey:SiswaID;references:ID" json:"siswa,omitempty"`
	KelasID       uint           `gorm:"not null;column:kelas_id" json:"kelas_id"`
	Kelas         Kelas          `gorm:"foreignKey:KelasID;references:ID" json:"kelas,omitempty"`
	MapelID       *uint          `gorm:"column:mapel_id" json:"mapel_id,omitempty"`
	MataPelajaran *MataPelajaran `gorm:"foreignKey:MapelID;references:ID" json:"mata_pelajaran,omitempty"`
	GuruID        uint           `gorm:"not null;column:guru_id" json:"guru_id"`
	Guru          Guru           `gorm:"foreignKey:GuruID;references:ID" json:"guru,omitempty"`
	TipeAbsensi   string         `gorm:"type:enum('kelas','mapel');not null" json:"tipe_absensi"`
	Tanggal       time.Time      `gorm:"type:date;not null" json:"tanggal"`
	Status        string         `gorm:"type:enum('masuk','izin','sakit','terlambat','alpa');not null" json:"status"`
	Keterangan    string         `gorm:"type:text" json:"keterangan,omitempty"`
	Semester      string         `gorm:"type:enum('ganjil','genap');not null" json:"semester"`
	TahunAjaran   string         `gorm:"type:varchar(9);not null" json:"tahun_ajaran"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type AbsensiResult struct {
	ID          uint    `json:"id"`
	SiswaID     uint    `json:"siswa_id"`
	NamaSiswa   string  `json:"nama_siswa"`
	KelasID     uint    `json:"kelas_id"`
	MapelID     *uint   `json:"mapel_id,omitempty"`
	NamaMapel   *string `json:"nama_mapel,omitempty"`
	GuruID      uint    `json:"guru_id"`
	TipeAbsensi string  `json:"tipe_absensi"`
	Tanggal     string  `json:"tanggal"`
	Status      string  `json:"status"`
	Keterangan  string  `json:"keterangan,omitempty"`
	TahunAjaran string  `json:"tahun_ajaran"`
	Semester    string  `json:"semester"`
}
