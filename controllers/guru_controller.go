package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllGurus(c *gin.Context) {
	var gurus []models.Guru
	if err := database.DB.Preload("GuruRoles").Find(&gurus).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data guru")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar guru berhasil diambil", gurus)
}

func GetGuruByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var guru models.Guru
	if err := database.DB.First(&guru, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "data guru tidak ditemukan")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "berhasil mengambil data guru", guru)
}

func CreateGuru(c *gin.Context) {
	var req requests.CreateGuruRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var existingGuru models.Guru
	if err := database.DB.Where("nip = ? OR email = ?", req.NIP, req.Email).First(&existingGuru).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "NIP/Email sudah terdaftar")
		return
	}

	plainPassword := req.NIP

	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi password")
		return
	}

	newGuru := models.Guru{
		Nama:         req.Nama,
		NIP:          req.NIP,
		NIK:          req.NIK,
		Email:        req.Email,
		Telepon:      req.Telepon,
		Alamat:       req.Alamat,
		JenisKelamin: req.JenisKelamin,
		Password:     hashedPassword,
	}

	if err := database.DB.Create(&newGuru).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat guru")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Guru berhasil dibuat", gin.H{
		"guru": newGuru,
	})
}

func UpdateGuru(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var guru models.Guru
	if err := database.DB.First(&guru, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "data guru tidak ditemukan")
		return
	}

	var req models.Guru
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	if req.JenisKelamin != "L" && req.JenisKelamin != "P" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jeniskelamin harus 'L' atau 'P'"})
		return
	}

	guru.Nama = req.Nama
	guru.NIP = req.NIP
	guru.NIK = req.NIK
	guru.Email = req.Email
	guru.Telepon = req.Telepon
	guru.Alamat = req.Alamat
	guru.JenisKelamin = req.JenisKelamin

	if err := database.DB.Save(&guru).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal memperbarui data guru")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "berhasil memperbarui data guru", guru)
}

func DeleteGuru(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	if err := database.DB.Delete(&models.Guru{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menghapus data guru")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "berhasil menghapus data guru", nil)
}

func GetPengajaranGuru(c *gin.Context) {
	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "role tidak ditemukan di context")
		return
	}
	role, _ := roleVal.(string)

	var guruID uint
	if role == "guru" {
		uidVal, ok := c.Get("user_id")
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "user_id tidak ditemukan di context")
			return
		}
		switch v := uidVal.(type) {
		case uint:
			guruID = v
		case int:
			guruID = uint(v)
		case int64:
			guruID = uint(v)
		case float64:
			guruID = uint(v)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "format user_id tidak dikenali")
			return
		}
	} else if role == "admin" {
		gq := c.Query("guru_id")
		if gq == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "admin harus menyertakan query param guru_id")
			return
		}
		id64, err := strconv.ParseUint(gq, 10, 64)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "guru_id tidak valid")
			return
		}
		guruID = uint(id64)
	} else {
		utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak")
		return
	}

	ta := c.DefaultQuery("tahun_ajaran", getTahunAjaranNow())
	sem := c.DefaultQuery("semester", getSemesterNow())

	db := database.DB

	var rows []requests.PengajaranRow
	err := db.Table("guru_mapel_kelas").
		Select("guru_mapel_kelas.mapel_id, mata_pelajarans.nama as mapel_nama, guru_mapel_kelas.kelas_id, kelas.nama as kelas_nama, guru_mapel_kelas.tahun_ajaran, guru_mapel_kelas.semester").
		Joins("JOIN mata_pelajarans ON mata_pelajarans.id = guru_mapel_kelas.mapel_id").
		Joins("JOIN kelas ON kelas.id = guru_mapel_kelas.kelas_id").
		Where("guru_mapel_kelas.guru_id = ? AND guru_mapel_kelas.tahun_ajaran = ? AND guru_mapel_kelas.semester = ?", guruID, ta, sem).
		Order("kelas.nama, mata_pelajarans.nama").
		Scan(&rows).Error

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data pengajaran: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar mapel & kelas yang diajar", rows)
}