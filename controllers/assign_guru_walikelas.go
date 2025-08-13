package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

	var kelas models.Kelas
	if err := tx.Preload("WaliKelas").First(&kelas, req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas")
		return
	}

	tx.Commit()
	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil ditetapkan", kelas)
}

func UnassignWaliKelas(c *gin.Context) {
	var req requests.UnassignWaliKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tx := database.DB.Begin()

	var kelas models.Kelas
	if err := tx.Preload("WaliKelas").First(&kelas, req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
		return
	}

	if kelas.WaliKelasID == nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "Kelas tersebut belum memiliki wali kelas")
		return
	}

	waliID := *kelas.WaliKelasID

	if err := tx.Model(&models.Kelas{}).
		Where("id = ?", req.KelasID).
		Update("wali_kelas_id", nil).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus wali kelas dari tabel kelas")
		return
	}

	if err := tx.Where("guru_id = ? AND kelas_id = ? AND role = 'wali_kelas'", waliID, req.KelasID).
		Delete(&models.GuruRole{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus role wali_kelas di guru_roles")
		return
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyelesaikan transaksi")
		return
	}

	kelas.WaliKelas = nil
	kelas.WaliKelasID = nil

	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil dihapus dari kelas", kelas)
}
