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

func GetKelasWali(c *gin.Context) {
	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "role tidak ditemukan di context")
		return
	}
	role, _ := roleVal.(string)

	db := database.DB

	if role == "wali_kelas" {
		uidVal, ok := c.Get("user_id")
		if !ok {
			utils.ErrorResponse(c, http.StatusUnauthorized, "user_id tidak ditemukan di context")
			return
		}
		var userID uint
		switch v := uidVal.(type) {
		case uint:
			userID = v
		case int:
			userID = uint(v)
		case int64:
			userID = uint(v)
		case float64:
			userID = uint(v)
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, "format user_id tidak dikenali")
			return
		}

		var kelas models.Kelas
		if err := db.Preload("WaliKelas").Where("wali_kelas_id = ?", userID).Find(&kelas).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas: "+err.Error())
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "Kelas yang Anda wali", kelas)
		return
	}

	if role == "admin" {
		q := c.Query("kelas_id")
		if q == "" {

			var kelasList []models.Kelas
			if err := db.Preload("WaliKelas").Find(&kelasList).Error; err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil daftar kelas: "+err.Error())
				return
			}
			utils.SuccessResponse(c, http.StatusOK, "Daftar kelas", kelasList)
			return
		}
		id64, err := strconv.ParseUint(q, 10, 64)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id tidak valid")
			return
		}
		var kelas models.Kelas
		if err := db.Preload("WaliKelas").First(&kelas, uint(id64)).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Detail kelas", kelas)
		return
	}

	utils.ErrorResponse(c, http.StatusForbidden, "akses ditolak")
}
