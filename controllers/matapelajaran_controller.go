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

func CreateMataPelajaran(c *gin.Context) {
	var req requests.CreateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	newMapel := models.MataPelajaran{
		Nama:    req.Nama,
		Kode:    req.Kode,
		Tingkat: req.Tingkat,
	}

	if err := database.DB.Create(&newMapel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat mata pelajaran")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Mata pelajaran berhasil dibuat", newMapel)
}

func GetAllMataPelajaran(c *gin.Context) {
	var mapel []models.MataPelajaran
	if err := database.DB.Find(&mapel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data mata pelajaran")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Data mata pelajaran berhasil diambil", mapel)
}

func GetMataPelajaranByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var mapel models.MataPelajaran
	if err := database.DB.First(&mapel, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data mata pelajaran tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data mata pelajaran berhasil ditemukan", mapel)
}

func UpdateMataPelajaran(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var mapel models.MataPelajaran
	if err := database.DB.First(&mapel, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data mata pelajaran tidak ditemukan")
		return
	}

	var req requests.CreateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	mapel.Nama = req.Nama
	mapel.Kode = req.Kode
	mapel.Tingkat = req.Tingkat

	if err := database.DB.Save(&mapel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data mata pelajaran")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data mata pelajaran berhasil diperbarui", mapel)
}

func DeleteMataPelajaran(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	if err := database.DB.Delete(&models.MataPelajaran{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data mata pelajaran")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data mata pelajaran berhasil dihapus", nil)
}
