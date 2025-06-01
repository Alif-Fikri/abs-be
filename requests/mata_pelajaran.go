package requests

type CreateMapelRequest struct {
	Nama        string `json:"nama" binding:"required"`
	Kode        string `json:"kode" binding:"required"`
	Tingkat     string `json:"tingkat" binding:"required,oneof=SD SMP SMA"`
	TahunAjaran string `json:"tahun_ajaran"`
	Semester    string `json:"semester" binding:"oneof=ganjil genap"`
	Hari        string `json:"hari" binding:"oneof=Senin Selasa Rabu Kamis Jumat Sabtu"`
	JamMulai    string `json:"jam_mulai" binding:"required"` // format: 08:00
	JamSelesai  string `json:"jam_selesai" binding:"required"`
}
