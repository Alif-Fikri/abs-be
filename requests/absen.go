package requests

type AbsensiRequest struct {
	SiswaID     uint   `json:"siswa_id" binding:"required"`
	KelasID     uint   `json:"kelas_id" binding:"required"`
	MapelID     *uint  `json:"mapel_id"` // opsional tergantung tipe_absensi
	GuruID      uint   `json:"guru_id" binding:"required"`
	TipeAbsensi string `json:"tipe_absensi" binding:"required,oneof=kelas mapel"`
	Tanggal     string `json:"tanggal" binding:"required"` // format: YYYY-MM-DD
	Status      string `json:"status" binding:"required,oneof=masuk izin sakit terlambat alpa"`
	Keterangan  string `json:"keterangan" binding:"omitempty"`
}
