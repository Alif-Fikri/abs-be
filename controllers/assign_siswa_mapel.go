package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"abs-be/database"
	"abs-be/firebaseclient"
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

	go func(siswa models.Siswa, mapel models.MataPelajaran) {
		var recipientIDs []uint

		recipientIDs = append(recipientIDs, siswa.ID)

		var guruIDs []uint
		rows, err := database.DB.Raw("SELECT guru_id FROM guru_mapels WHERE mapel_id = ?", mapel.ID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
				guruIDs = append(guruIDs, id)
			}
		}
		recipientIDs = append(recipientIDs, guruIDs...)

		var adminIDs []uint
		rowsAdmin, err := database.DB.Raw("SELECT id FROM admins").Rows()
		if err == nil {
			defer rowsAdmin.Close()
			for rowsAdmin.Next() {
				var id uint
				_ = rowsAdmin.Scan(&id)
				adminIDs = append(adminIDs, id)
			}
		}
		recipientIDs = append(recipientIDs, adminIDs...)

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

		title := "Penambahan Siswa ke Mapel"
		body := fmt.Sprintf("Siswa %s berhasil ditambahkan ke mata pelajaran %s", siswa.Nama, mapel.Nama)
		payload := map[string]interface{}{
			"type":       "assign_siswa_mapel",
			"siswa_id":   fmt.Sprintf("%d", siswa.ID),
			"siswa_nama": siswa.Nama,
			"mapel_id":   fmt.Sprintf("%d", mapel.ID),
			"mapel_nama": mapel.Nama,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "assign_siswa_mapel", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (assign siswa mapel): %v", err)
		}
	}(siswa, mapel)

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
	if tx.Error != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memulai transaksi: "+tx.Error.Error())
		return
	}
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

	var mapel models.MataPelajaran
	if err := database.DB.First(&mapel, req.MapelID).Error; err != nil {

		log.Printf("warning: gagal ambil data mapel untuk notifikasi: %v", err)
	}

	go func(siswa models.Siswa, mapel models.MataPelajaran, mapelID uint) {
		var recipientIDs []uint

		recipientIDs = append(recipientIDs, siswa.ID)

		var guruIDs []uint
		rows, err := database.DB.Raw("SELECT DISTINCT guru_id FROM guru_mapel_kelas WHERE mapel_id = ?", mapelID).Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
				guruIDs = append(guruIDs, id)
			}
		} else {
			log.Printf("warning: gagal ambil guru pengampu untuk mapel %d: %v", mapelID, err)
		}
		recipientIDs = append(recipientIDs, guruIDs...)

		var adminIDs []uint
		rowsAdmin, err := database.DB.Raw("SELECT id FROM admins").Rows()
		if err == nil {
			defer rowsAdmin.Close()
			for rowsAdmin.Next() {
				var id uint
				_ = rowsAdmin.Scan(&id)
				adminIDs = append(adminIDs, id)
			}
		} else {
			log.Printf("warning: gagal ambil admin ids: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)

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

		title := "Penghapusan dari Mata Pelajaran"
		mapelNama := mapel.Nama
		if mapelNama == "" {
			mapelNama = fmt.Sprintf("ID %d", mapelID)
		}
		body := fmt.Sprintf("Siswa %s telah dihapus dari mata pelajaran %s.", siswa.Nama, mapelNama)

		payload := map[string]interface{}{
			"type":       "unassign_siswa_mapel",
			"siswa_id":   fmt.Sprintf("%d", siswa.ID),
			"siswa_nama": siswa.Nama,
			"mapel_id":   fmt.Sprintf("%d", mapelID),
			"mapel_nama": mapelNama,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "unassign_siswa_mapel", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (unassign siswa mapel): %v", err)
		}
	}(siswa, mapel, req.MapelID)

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusOK, "Siswa dihapus dari mapel (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Siswa berhasil dihapus dari mapel", siswa)
}
