package requests

type KelasResponse struct {
	ID            uint   `json:"id"`
	Nama          string `json:"nama"`
	Tingkat       string `json:"tingkat"`
	WaliKelasID   *uint  `json:"wali_kelas_id,omitempty"`
	WaliKelasNama string `json:"wali_kelas_nama,omitempty"`
}

type MapelResponse struct {
	ID       uint   `json:"id"`
	Nama     string `json:"nama"`
	Kode     string `json:"kode"`
	Tingkat  string `json:"tingkat"`
	GuruID   *uint  `json:"guru_id,omitempty"`
	GuruNama string `json:"guru_nama,omitempty"`
}

type GuruMapelKelasResponse struct {
	ID        uint   `json:"id"`
	GuruID    uint   `json:"guru_id"`
	GuruNama  string `json:"guru_nama"`
	MapelID   uint   `json:"mapel_id"`
	MapelNama string `json:"mapel_nama"`
	KelasID   uint   `json:"kelas_id"`
	KelasNama string `json:"kelas_nama"`
}

type GuruResponse struct {
	ID           uint   `json:"id"`
	Nama         string `json:"nama"`
	NIP          string `json:"nip"`
	NIK          string `json:"nik"`
	Email        string `json:"email"`
	Telepon      string `json:"telepon"`
	Alamat       string `json:"alamat"`
	JenisKelamin string `json:"jenis_kelamin"`
}

type LoginResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Nama  string `json:"nama"`
	Role  string `json:"role"`
	Token string `json:"token"`
}

type SiswaResponse struct {
	ID           uint   `json:"id"`
	Nama         string `json:"nama"`
	NISN         string `json:"nisn"`
	TempatLahir  string `json:"tempat_lahir"`
	TanggalLahir string `json:"tanggal_lahir"` // format: YYYY-MM-DD
	JenisKelamin string `json:"jenis_kelamin"`
	NamaAyah     string `json:"nama_ayah"`
	NamaIbu      string `json:"nama_ibu"`
	Alamat       string `json:"alamat"`
	Agama        string `json:"agama"`
	Email        string `json:"email"`
	Telepon      string `json:"telepon"`
	AsalSekolah  string `json:"asal_sekolah"`
	KelasID      uint   `json:"kelas_id"`
}
