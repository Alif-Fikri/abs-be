package controllers

import (
	"fmt"
	"net/http"
	"time"

	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func CreateAbsensiSiswa(c *gin.Context) {
	var req requests.AbsensiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tanggal, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal harus YYYY-MM-DD")
		return
	}

	role := c.MustGet("role").(string)

	switch role {
	case "guru":

		if req.TipeAbsensi != "mapel" {
			utils.ErrorResponse(c, http.StatusForbidden, "Guru hanya dapat mengisi absen mapel")
			return
		}
	case "wali_kelas":

		if req.TipeAbsensi != "kelas" {
			utils.ErrorResponse(c, http.StatusForbidden, "Wali kelas hanya dapat mengisi absen kelas")
			return
		}
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Role tidak diizinkan")
		return
	}

	if req.TipeAbsensi == "mapel" && req.MapelID == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id harus diisi untuk absen mapel")
		return
	}

	absensi := models.AbsensiSiswa{
		SiswaID:     req.SiswaID,
		KelasID:     req.KelasID,
		MapelID:     req.MapelID,
		GuruID:      req.GuruID,
		TipeAbsensi: req.TipeAbsensi,
		Tanggal:     tanggal,
		Status:      req.Status,
		Keterangan:  req.Keterangan,
		TahunAjaran: getTahunAjaranNow(),
		Semester:    getSemesterNow(),
	}

	if err := database.DB.Create(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan absensi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil disimpan", absensi)
}

func ListStudentsForMapel(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	if mapelID == "" || kelasID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id & kelas_id wajib")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelasID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk mapel", siswa)
}

func ListStudentsForKelas(c *gin.Context) {
	guruID := c.MustGet("user_id").(uint)

	var kelas models.Kelas
	if err := database.DB.Where("wali_kelas_id = ?", guruID).First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Anda belum ditetapkan wali kelas")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelas.ID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk kelas Anda", siswa)
}

func RecapAbsensiMapel(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if mapelID == "" || kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id, kelas_id & tanggal wajib")
		return
	}
	tanggal, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	var recaps []struct {
		SiswaID uint
		Nama    string
		Status  string
	}
	if err := database.DB.
		Table("absensi_siswas").
		Select("siswa_id, siswas.nama, status").
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Where("tipe_absensi = ? AND mapel_id = ? AND kelas_id = ? AND tanggal = ?", "mapel", mapelID, kelasID, tanggal).
		Scan(&recaps).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi mapel")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rekap absensi mapel", recaps)
}

func RecapAbsensiKelas(c *gin.Context) {
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id & tanggal wajib")
		return
	}
	tanggal, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	var recaps []struct {
		SiswaID uint
		Nama    string
		Status  string
	}
	if err := database.DB.
		Table("absensi_siswas").
		Select("siswa_id, siswas.nama, status").
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Where("tipe_absensi = ? AND kelas_id = ? AND tanggal = ?", "kelas", kelasID, tanggal).
		Scan(&recaps).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rekap absensi kelas", recaps)
}

func getTahunAjaranNow() string {
	now := time.Now()
	year := now.Year()
	if now.Month() >= 7 {
		return fmt.Sprintf("%d/%d", year, year+1)
	}
	return fmt.Sprintf("%d/%d", year-1, year)
}

func getSemesterNow() string {
	m := time.Now().Month()
	if m >= 7 && m <= 12 {
		return "ganjil"
	}
	return "genap"
}
