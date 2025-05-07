package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateSiswa(c *gin.Context) {
	var req requests.CreateSiswaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var existingSiswa models.Siswa
	if err := database.DB.Where("nisn = ?", req.NISN).First(&existingSiswa).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "NISN sudah terdaftar")
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan format YYYY-MM-DD")
		return
	}

	newSiswa := models.Siswa{
		Nama:         req.Nama,
		NISN:         req.NISN,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		NamaAyah:     req.NamaAyah,
		NamaIbu:      req.NamaIbu,
		Alamat:       req.Alamat,
		Agama:        req.Agama,
		Email:        req.Email,
		Telepon:      req.Telepon,
		AsalSekolah:  req.AsalSekolah,
		KelasID:      req.KelasID,
	}

	if err := database.DB.Create(&newSiswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Siswa berhasil dibuat", newSiswa)
}

func GetSiswaByKelas(c *gin.Context) {
	kelasID := c.Param("kelasId")

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelasID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa berhasil diambil", siswa)
}
