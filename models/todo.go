package models

type Todo struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    UserID     uint      `json:"user_id"`
    Role       string    `json:"role"`
    Tanggal    string    `json:"tanggal"` // YYYY-MM-DD
    Deskripsi  string    `json:"deskripsi"`
    IsDone     bool      `json:"is_done"`
    JamDibuat  string    `json:"jam_dibuat"` // HH:MM:SS
    CreatedAt  string    `json:"created_at"`
    UpdatedAt  string    `json:"updated_at"`
}
