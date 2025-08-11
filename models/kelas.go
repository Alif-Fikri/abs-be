package models

import (
	"time"
)

type Kelas struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Nama        string    `gorm:"type:varchar(50);not null" json:"nama"`
	Tingkat     string    `gorm:"type:enum('SD','SMP','SMA');not null" json:"tingkat"`
	TahunAjaran string    `gorm:"type:varchar(9);not null" json:"tahun_ajaran"`
	WaliKelasID *uint     `json:"wali_kelas_id,omitempty"`
	WaliKelas   *Guru     `gorm:"foreignKey:WaliKelasID" json:"wali_kelas,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
