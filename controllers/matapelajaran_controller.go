package controllers

import (
	"abs-be/database"
	"abs-be/firebaseclient"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateMataPelajaran(c *gin.Context) {
	var req requests.CreateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	layout := "15:04"
	if _, err := time.Parse(layout, req.JamMulai); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format jam mulai salah, gunakan HH:MM")
		return
	}
	if _, err := time.Parse(layout, req.JamSelesai); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format jam selesai salah, gunakan HH:MM")
		return
	}

	jamMulaiStr := req.JamMulai + ":00"
	jamSelesaiStr := req.JamSelesai + ":00"

	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memulai transaksi: "+tx.Error.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi panic")
		}
	}()

	newMapel := models.MataPelajaran{
		Nama:       req.Nama,
		Kode:       req.Kode,
		Tingkat:    req.Tingkat,
		Semester:   req.Semester,
		Hari:       req.Hari,
		JamMulai:   jamMulaiStr,
		JamSelesai: jamSelesaiStr,
		IsActive:   req.IsActive,
	}

	if err := tx.Create(&newMapel).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat mata pelajaran: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit: "+err.Error())
		return
	}

	if err := database.DB.First(&newMapel, newMapel.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("warning preload mapel setelah create: %v", err)
	}

	go func(mapel models.MataPelajaran) {
		var recipientIDs []uint

		var adminIDs []uint
		if rows, err := database.DB.Raw("SELECT id FROM admins").Rows(); err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
				adminIDs = append(adminIDs, id)
			}
		} else {
			log.Printf("warning: gagal ambil admins: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)
		// notif semua guru
		// var guruIDs []uint
		// if rows, err := database.DB.Raw("SELECT id FROM gurus").Rows(); err == nil {
		// 	defer rows.Close()
		// 	for rows.Next() {
		// 		var id uint
		// 		_ = rows.Scan(&id)
		// 		guruIDs = append(guruIDs, id)
		// 	}
		// } else {
		// 	log.Printf("warning: gagal ambil gurus: %v", err)
		// }
		// recipientIDs = append(recipientIDs, guruIDs...)

		uniq := make(map[uint]struct{})
		finalRecipients := make([]uint, 0, len(recipientIDs))
		for _, id := range recipientIDs {
			if id == 0 {
				continue
			}
			if _, ok := uniq[id]; !ok {
				uniq[id] = struct{}{}
				finalRecipients = append(finalRecipients, id)
			}
		}

		title := "Mata Pelajaran Baru"
		body := fmt.Sprintf("Mata pelajaran %s (%s) berhasil dibuat.", mapel.Nama, mapel.Kode)
		payload := map[string]interface{}{
			"type":        "create_mapel",
			"mapel_id":    fmt.Sprintf("%d", mapel.ID),
			"mapel_nama":  mapel.Nama,
			"mapel_kode":  mapel.Kode,
			"tingkat":     mapel.Tingkat,
			"semester":    mapel.Semester,
			"hari":        mapel.Hari,
			"jam_mulai":   mapel.JamMulai,
			"jam_selesai": mapel.JamSelesai,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "create_mapel", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (create_mapel): %v", err)
		}
	}(newMapel)

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

	layout := "15:04"
	if _, err := time.Parse(layout, req.JamMulai); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format jam mulai salah, gunakan HH:MM")
		return
	}
	if _, err := time.Parse(layout, req.JamSelesai); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format jam selesai salah, gunakan HH:MM")
		return
	}

	jamMulaiStr := req.JamMulai + ":00"
	jamSelesaiStr := req.JamSelesai + ":00"

	mapel.Nama = req.Nama
	mapel.Kode = req.Kode
	mapel.Tingkat = req.Tingkat
	mapel.Semester = req.Semester
	mapel.Hari = req.Hari
	mapel.JamMulai = jamMulaiStr
	mapel.JamSelesai = jamSelesaiStr
	mapel.IsActive = req.IsActive

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

	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memulai transaksi: "+tx.Error.Error())
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Terjadi panic")
		}
	}()

	var mapel models.MataPelajaran
	if err := tx.First(&mapel, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Mata pelajaran tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data mapel: "+err.Error())
		return
	}

	var guruIDs []uint
	if rows, err := tx.Raw("SELECT DISTINCT guru_id FROM guru_mapel_kelas WHERE mapel_id = ?", mapel.ID).Rows(); err == nil {
		defer rows.Close()
		for rows.Next() {
			var id uint
			_ = rows.Scan(&id)
			guruIDs = append(guruIDs, id)
		}
	} else {
		log.Printf("warning: gagal ambil guru pengampu: %v", err)
	}

	var siswaIDs []uint
	if err := tx.Table("mapel_siswas").Where("mapel_id = ?", mapel.ID).Pluck("siswa_id", &siswaIDs).Error; err != nil {
		log.Printf("warning: gagal ambil siswa di mapel: %v", err)
	}

	if err := tx.Delete(&mapel).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus mata pelajaran: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit: "+err.Error())
		return
	}

	go func(mapel models.MataPelajaran, guruIDs, siswaIDs []uint) {
		var recipientIDs []uint

		var adminIDs []uint
		if rows, err := database.DB.Raw("SELECT id FROM admins").Rows(); err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
				adminIDs = append(adminIDs, id)
			}
		} else {
			log.Printf("warning: gagal ambil admins: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)
		recipientIDs = append(recipientIDs, guruIDs...)
		recipientIDs = append(recipientIDs, siswaIDs...)

		uniq := make(map[uint]struct{})
		finalRecipients := make([]uint, 0, len(recipientIDs))
		for _, id := range recipientIDs {
			if id == 0 {
				continue
			}
			if _, ok := uniq[id]; !ok {
				uniq[id] = struct{}{}
				finalRecipients = append(finalRecipients, id)
			}
		}

		title := "Mata Pelajaran Dihapus"
		body := fmt.Sprintf("Mata pelajaran %s (%s) telah dihapus.", mapel.Nama, mapel.Kode)
		payload := map[string]interface{}{
			"type":       "delete_mapel",
			"mapel_id":   fmt.Sprintf("%d", mapel.ID),
			"mapel_nama": mapel.Nama,
			"mapel_kode": mapel.Kode,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "delete_mapel", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (delete_mapel): %v", err)
		}
	}(mapel, guruIDs, siswaIDs)

	utils.SuccessResponse(c, http.StatusOK, "Data mata pelajaran berhasil dihapus", nil)
}