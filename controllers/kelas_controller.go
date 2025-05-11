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

func CreateKelas(c *gin.Context) {
	var req requests.CreateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	newKelas := models.Kelas{
		Nama:        req.Nama,
		Tingkat:     req.Tingkat,
		TahunAjaran: req.TahunAjaran,
	}

	if err := database.DB.Create(&newKelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Kelas berhasil dibuat", newKelas)
}

func GetAllKelas(c *gin.Context) {
	var kelas []models.Kelas
	if err := database.DB.Preload("WaliKelas").Find(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar kelas berhasil diambil", kelas)
}

func GetKelasByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.Preload("WaliKelas").First(&kelas, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data kelas tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data kelas", kelas)
}

func UpdateKelas(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.First(&kelas, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data kelas tidak ditemukan")
		return
	}

	var req requests.CreateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	kelas.Nama = req.Nama
	kelas.Tingkat = req.Tingkat
	kelas.TahunAjaran = req.TahunAjaran

	if err := database.DB.Save(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data kelas berhasil diperbarui", kelas)
}

func DeleteKelas(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	if err := database.DB.Delete(&models.Kelas{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data kelas berhasil dihapus", nil)
}
