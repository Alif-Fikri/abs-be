package requests

type CreateKelasRequest struct {
	Nama        string `json:"nama" binding:"required"`
	Tingkat     string `json:"tingkat" binding:"required,oneof=SD SMP SMA"`
	TahunAjaran string `json:"tahun_ajaran" binding:"required"`
}
