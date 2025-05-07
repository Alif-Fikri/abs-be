package requests

type CreateGuruRequest struct {
	Nama         string `json:"nama" binding:"required"`
	NIP          string `json:"nip" binding:"required"`
	NIK          string `json:"nik" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Telepon      string `json:"telepon" binding:"omitempty"`
	Alamat       string `json:"alamat" binding:"required"`
	JenisKelamin string `json:"jenis_kelamin" binding:"required,oneof=L P"`
	Password     string `json:"password" binding:"required"`
}

type GuruMapelKelasRequest struct {
	GuruID  uint `json:"guru_id" binding:"required"`
	MapelID uint `json:"mapel_id" binding:"required"`
	KelasID uint `json:"kelas_id" binding:"required"`
}

type AssignWaliKelasRequest struct {
	GuruID  uint `json:"guru_id" binding:"required"`
	KelasID uint `json:"kelas_id" binding:"required"`
}
