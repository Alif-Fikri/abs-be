package models

import "time"

type Todo struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    UserID     uint      `json:"user_id"`
    Role       string    `json:"role" gorm:"type:enum('admin','guru','wali_kelas');not null"`
    Tanggal    time.Time `json:"tanggal" gorm:"type:date"`
    Deskripsi  string    `json:"deskripsi"`
    IsDone     bool      `json:"is_done" gorm:"default:false"`
    JamDibuat  string    `json:"jam_dibuat" gorm:"type:time"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}