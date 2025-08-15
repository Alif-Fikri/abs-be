package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"
	"strconv"
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

	plainPassword := tglLahir.Format("02012006")

	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi password")
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
		Password:     hashedPassword,
	}

	if err := database.DB.Create(&newSiswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Siswa berhasil dibuat", gin.H{
		"siswa": newSiswa,
	})
}

func GetSiswaByKelas(c *gin.Context) {
	kelasIDParam := c.Param("id")
	kid, err := strconv.ParseUint(kelasIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelasId tidak valid")
		return
	}
	kidUint := uint(kid)

	db := database.DB

	var ids []uint
	if err := db.
		Table("siswas").
		Select("DISTINCT siswas.id").
		Joins("LEFT JOIN kelas_siswas ks ON ks.siswa_id = siswas.id").
		Where("siswas.kelas_id = ? OR ks.kelas_id = ?", kidUint, kidUint).
		Pluck("siswas.id", &ids).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil daftar siswa: "+err.Error())
		return
	}

	if len(ids) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Tidak ada siswa di kelas tersebut")
		return
	}

	var siswa []models.Siswa
	if err := db.Preload("Kelas").Preload("MataPelajaran").Find(&siswa, ids).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memuat relasi siswa: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa berhasil diambil", siswa)
}

func GetAllSiswa(c *gin.Context) {
	var siswa []models.Siswa
	if err := database.DB.Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data siswa")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil diambil", siswa)
}

func GetSiswaByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var siswa models.Siswa
	if err := database.DB.First(&siswa, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil ditemukan", siswa)
}

func UpdateSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var siswa models.Siswa
	if err := database.DB.First(&siswa, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	var req requests.CreateSiswaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan format YYYY-MM-DD")
		return
	}

	siswa.Nama = req.Nama
	siswa.NISN = req.NISN
	siswa.TempatLahir = req.TempatLahir
	siswa.TanggalLahir = tglLahir
	siswa.JenisKelamin = req.JenisKelamin
	siswa.NamaAyah = req.NamaAyah
	siswa.NamaIbu = req.NamaIbu
	siswa.Alamat = req.Alamat
	siswa.Agama = req.Agama
	siswa.Email = req.Email
	siswa.Telepon = req.Telepon
	siswa.AsalSekolah = req.AsalSekolah

	if err := database.DB.Save(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil diperbarui", siswa)
}

func DeleteSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	if err := database.DB.Delete(&models.Siswa{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil dihapus", nil)
}

func GetProfilSiswa(c *gin.Context) {
	siswaID := c.MustGet("user_id").(uint)

	var siswa models.Siswa
	if err := database.DB.First(&siswa, siswaID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	response := requests.SiswaResponse{
		ID:           siswa.ID,
		Nama:         siswa.Nama,
		NISN:         siswa.NISN,
		TempatLahir:  siswa.TempatLahir,
		TanggalLahir: siswa.TanggalLahir.Format("2006-01-02"),
		JenisKelamin: siswa.JenisKelamin,
		NamaAyah:     siswa.NamaAyah,
		NamaIbu:      siswa.NamaIbu,
		Alamat:       siswa.Alamat,
		Agama:        siswa.Agama,
		Email:        siswa.Email,
		Telepon:      siswa.Telepon,
		AsalSekolah:  siswa.AsalSekolah,
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil siswa", response)
}

func GetAbsensiSiswa(c *gin.Context) {
	siswaID := c.MustGet("user_id").(uint)

	tanggal := c.Query("tanggal")
	tipe := c.Query("tipe")

	query := database.DB.Where("siswa_id = ?", siswaID)

	if tanggal != "" {
		t, err := time.Parse("2006-01-02", tanggal)
		if err == nil {
			query = query.Where("tanggal = ?", t)
		}
	}

	if tipe != "" && (tipe == "kelas" || tipe == "mapel") {
		query = query.Where("tipe_absensi = ?", tipe)
	}

	var absensi []models.AbsensiSiswa
	if err := query.Find(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data absensi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data absensi", absensi)
}
