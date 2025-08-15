package controllers

import (
	"net/http"

	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func AssignSiswaToMapel(c *gin.Context) {
	var req requests.AssignSiswaMapelRequest
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
	var mapel models.MataPelajaran
	if err := tx.First(&mapel, req.MapelID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Mapel tidak ditemukan")
		return
	}

	var cnt int64
	if err := tx.Table("mapel_siswas").
		Where("siswa_id = ? AND mapel_id = ?", req.SiswaID, req.MapelID).
		Count(&cnt).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa data: "+err.Error())
		return
	}
	if cnt > 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Siswa sudah terdaftar untuk mapel ini")
		return
	}

	if err := tx.Exec("INSERT INTO mapel_siswas (siswa_id, mapel_id) VALUES (?, ?)", req.SiswaID, req.MapelID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menambahkan siswa ke mapel: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusCreated, "Siswa ditambahkan ke mapel (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Siswa berhasil ditambahkan ke mapel", siswa)
}

func UnassignSiswaFromMapel(c *gin.Context) {
	var req requests.UnassignSiswaMapelRequest
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
	if err := tx.Table("mapel_siswas").
		Where("siswa_id = ? AND mapel_id = ?", req.SiswaID, req.MapelID).
		Count(&cnt).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa data: "+err.Error())
		return
	}
	if cnt == 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak terdaftar di mapel tersebut")
		return
	}

	if err := tx.Exec("DELETE FROM mapel_siswas WHERE siswa_id = ? AND mapel_id = ?", req.SiswaID, req.MapelID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus hubungan siswa-mapel: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusOK, "Siswa dihapus dari mapel (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Siswa berhasil dihapus dari mapel", siswa)
}
