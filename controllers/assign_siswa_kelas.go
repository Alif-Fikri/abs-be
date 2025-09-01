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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&siswa, req.SiswaID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data siswa: "+err.Error())
		return
	}

	var kelas models.Kelas
	if err := tx.First(&kelas, req.KelasID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas: "+err.Error())
		return
	}

	if siswa.KelasID != nil && *siswa.KelasID != 0 && *siswa.KelasID != req.KelasID {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Siswa sudah memiliki kelas utama lain; hapus/ubah dulu sebelum assign ke kelas berbeda")
		return
	}

	var otherCnt int64
	if err := tx.Table("kelas_siswas").
		Where("siswa_id = ? AND kelas_id != ?", req.SiswaID, req.KelasID).
		Count(&otherCnt).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa pendaftaran kelas: "+err.Error())
		return
	}
	if otherCnt > 0 {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Siswa sudah terdaftar di kelas lain; hapus relasi sebelumnya sebelum menambahkan ke kelas ini")
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
		if siswa.KelasID == nil || *siswa.KelasID != req.KelasID {
			if err := tx.Model(&siswa).Update("kelas_id", req.KelasID).Error; err != nil {
				tx.Rollback()
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal update kelas utama siswa: "+err.Error())
				return
			}
		}
		if err := tx.Commit().Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
			return
		}

		_ = database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID)
		utils.SuccessResponse(c, http.StatusOK, "Siswa sudah terdaftar di kelas ini", siswa)
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	go func(s models.Siswa, k models.Kelas) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in notif goroutine: %v", r)
			}
		}()

		var recipientIDs []uint
		recipientIDs = append(recipientIDs, s.ID)

		if k.WaliKelasID != nil && *k.WaliKelasID != 0 {
			recipientIDs = append(recipientIDs, *k.WaliKelasID)
		}

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
			if _, ok := uniq[id]; !ok && id != 0 {
				uniq[id] = struct{}{}
				finalRecipients = append(finalRecipients, id)
			}
		}

		title := "Penambahan Siswa ke Kelas"
		body := fmt.Sprintf("Siswa %s telah ditambahkan ke kelas %s.", s.Nama, k.Nama)
		payload := map[string]interface{}{
			"type":       "assign_siswa_kelas",
			"siswa_id":   s.ID,
			"siswa_nama": s.Nama,
			"kelas_id":   k.ID,
			"kelas_nama": k.Nama,
			"actor":      "admin",
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "assign_siswa_kelas", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (assign siswa kelas): %v", err)
		}
	}(siswa, kelas)

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

	go func(siswa models.Siswa, kelas models.Kelas) {
		var recipientIDs []uint
		recipientIDs = append(recipientIDs, siswa.ID)

		if kelas.WaliKelasID != nil && *kelas.WaliKelasID != 0 {
			recipientIDs = append(recipientIDs, *kelas.WaliKelasID)
		}

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
			if _, ok := uniq[id]; !ok && id != 0 {
				uniq[id] = struct{}{}
				finalRecipients = append(finalRecipients, id)
			}
		}

		title := "Penghapusan Siswa dari Kelas"
		body := fmt.Sprintf("Siswa %s telah dihapus dari kelas %s.", siswa.Nama, kelas.Nama)
		payload := map[string]interface{}{
			"type":       "unassign_siswa_kelas",
			"siswa_id":   siswa.ID,
			"siswa_nama": siswa.Nama,
			"kelas_id":   kelas.ID,
			"kelas_nama": kelas.Nama,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "unassign_siswa_kelas", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (unassign siswa kelas): %v", err)
		}
	}(siswa, kelas)

	if err := database.DB.Preload("Kelas").Preload("MataPelajaran").First(&siswa, req.SiswaID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusOK, "Siswa dihapus dari kelas (tapi gagal memuat relasi)", siswa)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Siswa berhasil dihapus dari kelas", siswa)
}
