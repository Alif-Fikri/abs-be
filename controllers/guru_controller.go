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

