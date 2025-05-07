package requests

type CreateMapelRequest struct {
	Nama    string `json:"nama" binding:"required"`
	Kode    string `json:"kode" binding:"required"`
	Tingkat string `json:"tingkat" binding:"required,oneof=SD SMP SMA"`
}
