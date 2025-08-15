package controllers

import (
	"net/http"

	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func AssignSiswaToKelas(c *gin.Context) {
	var req requests.AssignSiswaKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	role, _ := c.Get("role")
	if roleStr, ok := role.(string); !ok || roleStr != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Hanya admin yang dapat melakukan aksi ini")
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var siswa models.Siswa
	if err := tx.First(&siswa, req.SiswaID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	var kelas models.Kelas
	if err := tx.First(&kelas, req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
		return
	}

	var cnt int64
	if err := tx.Table("kelas_siswas").
		Where("siswa_id = ? AND kelas_id = ?", req.SiswaID, req.KelasID).
		Count(&cnt).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa data: "+err.Error())
		return
	}
	if cnt > 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Siswa sudah terdaftar di kelas ini")
		return
	}

	if err := tx.Exec("INSERT INTO kelas_siswas (siswa_id, kelas_id) VALUES (?, ?)", req.SiswaID, req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menambahkan siswa ke kelas: "+err.Error())
		return
	}

	if err := tx.Model(&siswa).Update("kelas_id", req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update kelas utama siswa: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusCreated, "Siswa ditambahkan ke kelas (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Siswa berhasil ditambahkan ke kelas", siswa)
}

func UnassignSiswaFromKelas(c *gin.Context) {
	var req requests.UnassignSiswaKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	role, _ := c.Get("role")
	if roleStr, ok := role.(string); !ok || roleStr != "admin" {
		utils.ErrorResponse(c, http.StatusForbidden, "Hanya admin yang dapat melakukan aksi ini")
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var siswa models.Siswa
	if err := tx.First(&siswa, req.SiswaID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	var cnt int64
	if err := tx.Table("kelas_siswas").
		Where("siswa_id = ? AND kelas_id = ?", req.SiswaID, req.KelasID).
		Count(&cnt).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa data: "+err.Error())
		return
	}
	if cnt == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak terdaftar di kelas tersebut")
		return
	}

	if err := tx.Exec("DELETE FROM kelas_siswas WHERE siswa_id = ? AND kelas_id = ?", req.SiswaID, req.KelasID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus hubungan siswa-kelas: "+err.Error())
		return
	}

	if err := tx.Model(&siswa).Update("kelas_id", nil).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal reset kelas utama siswa: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusOK, "Siswa dihapus dari kelas (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Siswa berhasil dihapus dari kelas", siswa)
}
