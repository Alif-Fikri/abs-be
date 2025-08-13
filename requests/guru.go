package requests

type CreateGuruRequest struct {
	Nama         string `json:"nama" binding:"required"`
	NIP          string `json:"nip" binding:"required"`
	NIK          string `json:"nik" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Telepon      string `json:"telepon" binding:"omitempty"`
	Alamat       string `json:"alamat" binding:"required"`
	JenisKelamin string `json:"jenis_kelamin" binding:"required,oneof=L P"`
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

type UnassignWaliKelasRequest struct {
	KelasID uint `json:"kelas_id" binding:"required"`
}

type AssignMapelKelasRequest struct {
	GuruID      uint   `json:"guru_id" binding:"required"`
	MapelID     uint   `json:"mapel_id" binding:"required"`
	KelasID     uint   `json:"kelas_id" binding:"required"`
	TahunAjaran string `json:"tahun_ajaran" binding:"required,len=9"` // format: 2025/2026
	Semester    string `json:"semester" binding:"required,oneof=ganjil genap"`
}
