package requests

type AssignSiswaKelasRequest struct {
    SiswaID uint `json:"siswa_id" binding:"required"`
    KelasID uint `json:"kelas_id" binding:"required"`
}

type UnassignSiswaKelasRequest struct {
    SiswaID uint `json:"siswa_id" binding:"required"`
    KelasID uint `json:"kelas_id" binding:"required"`
}

type AssignSiswaMapelRequest struct {
    SiswaID uint `json:"siswa_id" binding:"required"`
    MapelID uint `json:"mapel_id" binding:"required"`
}

type UnassignSiswaMapelRequest struct {
    SiswaID uint `json:"siswa_id" binding:"required"`
    MapelID uint `json:"mapel_id" binding:"required"`
}
