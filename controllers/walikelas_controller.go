package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllWaliKelas(c *gin.Context) {
	var kelas []models.Kelas
	if err := database.DB.Preload("WaliKelas").Find(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal mengambil data wali kelas")
		return
	}

	var result []gin.H
	for _, k := range kelas {
		if k.WaliKelas != nil {
			result = append(result, gin.H{
				"kelas_id":        k.ID,
				"nama_kelas":      k.Nama,
				"tingkat":         k.Tingkat,
				"tahun_ajaran":    k.TahunAjaran,
				"wali_kelas_id":   k.WaliKelas.ID,
				"wali_kelas_nama": k.WaliKelas.Nama,
			})
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "berhasil mengambil data wali kelas", result)
}

func GetKelasWaliKelas(c *gin.Context) {
	guruID := c.MustGet("userID").(uint)

	var kelas models.Kelas
	if err := database.DB.Preload("WaliKelas").Where("wali_kelas_id = ?", guruID).First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "guru ini bukan wali kelas dari kelas manapun")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "guru ini adalah wali kelas", kelas)
}

func GetSiswaByWaliKelas(c *gin.Context) {
	guruIDParam := c.Param("guruID")
	guruID, err := strconv.Atoi(guruIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID guru tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.Where("wali_kelas_id = ?", guruID).First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan untuk wali kelas ini")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelas.ID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa berhasil diambil", siswa)
}

func UpdateWaliKelas(c *gin.Context) {
	kelasIDParam := c.Param("kelasID")
	kelasID, err := strconv.Atoi(kelasIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID kelas tidak valid")
		return
	}

	var req struct {
		GuruID uint `json:"guru_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var kelas models.Kelas
	if err := database.DB.First(&kelas, kelasID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
		return
	}

	kelas.WaliKelasID = &req.GuruID

	if err := database.DB.Save(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui wali kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil diperbarui", kelas)
}

func RemoveWaliKelas(c *gin.Context) {
	kelasIDParam := c.Param("kelasID")
	kelasID, err := strconv.Atoi(kelasIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID kelas tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.First(&kelas, kelasID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
		return
	}

	kelas.WaliKelasID = nil

	if err := database.DB.Save(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus wali kelas dari kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil dihapus dari kelas", kelas)
}
