package models

import "time"

type Todo struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AdminID     *uint     `json:"admin_id,omitempty"`
	Admin       *Admin    `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
	GuruID      *uint     `json:"guru_id,omitempty"`
	Guru        *Guru     `gorm:"foreignKey:GuruID" json:"guru,omitempty"`
	WaliKelasID *uint     `json:"wali_kelas_id,omitempty"`
	WaliKelas   *Guru     `gorm:"foreignKey:WaliKelasID" json:"wali_kelas,omitempty"`
	Role        string    `json:"role" gorm:"type:enum('admin','guru','wali_kelas');not null"`
	Tanggal     time.Time `json:"tanggal" gorm:"type:date"`
	Deskripsi   string    `json:"deskripsi"`
	IsDone      bool      `json:"is_done" gorm:"default:false"`
	JamDibuat   string    `json:"jam_dibuat" gorm:"type:time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
