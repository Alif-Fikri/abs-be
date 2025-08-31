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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi")
		return
	}

	go func(kelas models.Kelas) {
		var recipientIDs []uint
		if kelas.WaliKelasID != nil {
			recipientIDs = append(recipientIDs, *kelas.WaliKelasID)
		}

		var adminIDs []uint
		rows, err := database.DB.Raw("SELECT id FROM admins").Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
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

		title := "Penetapan Wali Kelas"
		body := fmt.Sprintf(
			"Guru %s telah resmi ditetapkan sebagai wali kelas %s.",
			kelas.WaliKelas.Nama,
			kelas.Nama,
		)

		payload := map[string]interface{}{
			"type":       "assign_wali_kelas",
			"guru_id":    fmt.Sprintf("%d", *kelas.WaliKelasID),
			"kelas_id":   fmt.Sprintf("%d", kelas.ID),
			"kelas_nama": kelas.Nama,
		}

		if err := firebaseclient.NotifyUsers(
			context.Background(),
			"assign_wali_kelas",
			title,
			body,
			payload,
			finalRecipients,
		); err != nil {
			log.Printf("NotifyUsers error (assign wali kelas): %v", err)
		}
	}(kelas)

	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil ditetapkan", kelas)
}

func UnassignWaliKelas(c *gin.Context) {
	var req requests.UnassignWaliKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
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

	var kelas models.Kelas
	if err := tx.Preload("WaliKelas").First(&kelas, req.KelasID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Kelas tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas: "+err.Error())
		return
	}

	if kelas.WaliKelasID == nil || kelas.WaliKelas == nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "Kelas tersebut belum memiliki wali kelas")
		return
	}

	prevWaliID := *kelas.WaliKelasID
	prevWaliNama := kelas.WaliKelas.Nama
	kelasID := kelas.ID
	kelasNama := kelas.Nama

	if err := tx.Model(&models.Kelas{}).
		Where("id = ?", kelasID).
		Update("wali_kelas_id", nil).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus wali kelas dari tabel kelas: "+err.Error())
		return
	}

	if err := tx.Where("guru_id = ? AND kelas_id = ? AND role = 'wali_kelas'", prevWaliID, kelasID).
		Delete(&models.GuruRole{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus role wali_kelas: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	go func(waliID uint, waliNama, kelasNama string, kelasID uint) {

		var recipientIDs []uint
		recipientIDs = append(recipientIDs, waliID)

		var adminIDs []uint
		rows, err := database.DB.Raw("SELECT id FROM admins").Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
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

		title := "Pencabutan Wali Kelas"
		body := fmt.Sprintf("Penetapan Wali Kelas untuk kelas %s dibatalkan. Guru %s tidak lagi menjabat wali kelas %s.", kelasNama, waliNama, kelasNama)
		payload := map[string]interface{}{
			"type":       "unassign_wali_kelas",
			"guru_id":    fmt.Sprintf("%d", waliID),
			"kelas_id":   fmt.Sprintf("%d", kelasID),
			"kelas_nama": kelasNama,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "unassign_wali_kelas", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (unassign wali kelas): %v", err)
		}
	}(prevWaliID, prevWaliNama, kelasNama, kelasID)

	kelas.WaliKelas = nil
	kelas.WaliKelasID = nil

	utils.SuccessResponse(c, http.StatusOK, "Wali kelas berhasil dihapus dari kelas", kelas)
}
