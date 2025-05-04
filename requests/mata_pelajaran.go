package requests

type MapelRequest struct {
	Nama    string `json:"nama" binding:"required"`
	Kode    string `json:"kode" binding:"required"`
	Tingkat string `json:"tingkat" binding:"required,oneof=SD SMP SMA"`
	GuruID  uint   `json:"guru_id" binding:"required"`
}
