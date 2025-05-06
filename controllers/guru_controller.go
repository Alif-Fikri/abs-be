package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllGurus(c *gin.Context) {
	var gurus []models.Guru
	if err := database.DB.Find(&gurus).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal mengambil data guru")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "berhasil mengambil semua data guru", gurus)
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
	var req models.Guru
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	if err := database.DB.Create(&req).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat data guru")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "berhasil membuat data guru", req)
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