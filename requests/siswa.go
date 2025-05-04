package requests

type SiswaRequest struct {
	Nama         string `json:"nama" binding:"required"`
	NISN         string `json:"nisn" binding:"required"`
	TempatLahir  string `json:"tempat_lahir" binding:"required"`
	TanggalLahir string `json:"tanggal_lahir" binding:"required"` // format: YYYY-MM-DD
	JenisKelamin string `json:"jenis_kelamin" binding:"required,oneof=L P"`
	NamaAyah     string `json:"nama_ayah" binding:"required"`
	NamaIbu      string `json:"nama_ibu" binding:"required"`
	Alamat       string `json:"alamat" binding:"required"`
	Agama        string `json:"agama" binding:"required"`
	Email        string `json:"email" binding:"omitempty,email"`
	Telepon      string `json:"telepon" binding:"omitempty"`
	AsalSekolah  string `json:"asal_sekolah" binding:"required"`
	KelasID      uint   `json:"kelas_id" binding:"required"`
}
