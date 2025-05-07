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

	hashedPassword, err := utils.HashPassword(req.Password)
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

	utils.SuccessResponse(c, http.StatusCreated, "Guru berhasil dibuat", newGuru)
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

	guru.Nama = req.Nama
	guru.NIP = req.NIP
	guru.NIK = req.NIK
	guru.Email = req.Email
	guru.Telepon = req.Telepon
	guru.Alamat = req.Alamat
	guru.JenisKelamin = req.JenisKelamin
	guru.Password = req.Password

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

func AssignWaliKelas(c *gin.Context) {
	var req requests.AssignWaliKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var existingRole models.GuruRole
	if err := database.DB.Where("guru_id = ? AND role = 'wali_kelas'", req.GuruID).First(&existingRole).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Guru sudah menjadi wali kelas di kelas lain")
		return
	}

	tx := database.DB.Begin()

	if err := tx.Model(&models.Kelas{}).
		Where("id = ?", req.KelasID).
		Update("wali_kelas_id", nil).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengupdate kelas")
		return
	}

	if err := tx.Model(&models.Kelas{}).
		Where("id = ?", req.KelasID).
		Update("wali_kelas_id", req.GuruID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menetapkan wali kelas")
		return
	}

	role := models.GuruRole{
		GuruID:  req.GuruID,
		Role:    "wali_kelas",
		KelasID: &req.KelasID,
	}

	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat role wali kelas")
		return
	}

	tx.Commit()
	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil ditetapkan", nil)
}
