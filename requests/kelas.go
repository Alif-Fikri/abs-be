package requests

type KelasRequest struct {
	Nama        string `json:"nama" binding:"required"`
	Tingkat     string `json:"tingkat" binding:"required,oneof=SD SMP SMA"`
	WaliKelasID uint   `json:"wali_kelas_id" binding:"omitempty"`
}
