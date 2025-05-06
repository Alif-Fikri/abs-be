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
				"wali_kelas_id":   k.WaliKelas.ID,
				"wali_kelas_nama": k.WaliKelas.Nama,
			})
		}
	}

	utils.SuccessResponse(c, http.StatusOK, "berhasil mengambil data wali kelas", result)
}

func GetKelasWaliKelas(c *gin.Context) {
	guruIDParam := c.Param("guruID")
	guruID, err := strconv.Atoi(guruIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID guru tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.Preload("WaliKelas").Where("wali_kelas_id = ?", guruID).First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Guru ini bukan wali kelas dari kelas manapun")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Guru adalah wali kelas", kelas)
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
