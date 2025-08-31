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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateKelas(c *gin.Context) {
	var req requests.CreateKelasRequest
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

	newKelas := models.Kelas{
		Nama:        req.Nama,
		Tingkat:     req.Tingkat,
		TahunAjaran: req.TahunAjaran,
	}

	if err := tx.Create(&newKelas).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat kelas: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit: "+err.Error())
		return
	}

	if err := database.DB.Preload("WaliKelas").First(&newKelas, newKelas.ID).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("warning: gagal preload kelas setelah create: %v", err)
	}

	go func(kelas models.Kelas) {
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
			log.Printf("warning: gagal ambil admin ids: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)

		// (notif semua guru)
		// var guruIDs []uint
		// if rows, err := database.DB.Raw("SELECT id FROM gurus").Rows(); err == nil {
		// 	defer rows.Close()
		// 	for rows.Next() {
		// 		var id uint
		// 		_ = rows.Scan(&id)
		// 		guruIDs = append(guruIDs, id)
		// 	}
		// } else {
		// 	log.Printf("warning: gagal ambil guru ids: %v", err)
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

		title := "Kelas Baru Dibuat"
		body := fmt.Sprintf("Kelas %s (%s, TA %s) telah dibuat.", kelas.Nama, kelas.Tingkat, kelas.TahunAjaran)
		payload := map[string]interface{}{
			"type":        "create_kelas",
			"kelas_id":    fmt.Sprintf("%d", kelas.ID),
			"kelas_nama":  kelas.Nama,
			"tingkat":     kelas.Tingkat,
			"tahun_ajaran": kelas.TahunAjaran,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "create_kelas", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (create_kelas): %v", err)
		}
	}(newKelas)

	utils.SuccessResponse(c, http.StatusCreated, "Kelas berhasil dibuat", newKelas)
}

func GetAllKelas(c *gin.Context) {
	var kelas []models.Kelas
	if err := database.DB.Preload("WaliKelas").Find(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar kelas berhasil diambil", kelas)
}

func GetKelasByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.Preload("WaliKelas").First(&kelas, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data kelas tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data kelas", kelas)
}

func UpdateKelas(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var kelas models.Kelas
	if err := database.DB.First(&kelas, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Data kelas tidak ditemukan")
		return
	}

	var req requests.CreateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	kelas.Nama = req.Nama
	kelas.Tingkat = req.Tingkat
	kelas.TahunAjaran = req.TahunAjaran

	if err := database.DB.Save(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data kelas berhasil diperbarui", kelas)
}

func DeleteKelas(c *gin.Context) {
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

	var kelas models.Kelas
	if err := tx.Preload("WaliKelas").First(&kelas, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Data kelas tidak ditemukan")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal ambil data kelas: "+err.Error())
		return
	}

	var siswaIDs []uint
	if err := tx.Table("kelas_siswas").Where("kelas_id = ?", kelas.ID).Pluck("siswa_id", &siswaIDs).Error; err != nil {
		log.Printf("warning: gagal ambil siswa di kelas %d: %v", kelas.ID, err)
	}

	if err := tx.Delete(&kelas).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data kelas: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	go func(kelas models.Kelas, siswaIDs []uint) {
		var recipientIDs []uint

		if kelas.WaliKelasID != nil && *kelas.WaliKelasID != 0 {
			recipientIDs = append(recipientIDs, *kelas.WaliKelasID)
		}

		var adminIDs []uint
		if rows, err := database.DB.Raw("SELECT id FROM admins").Rows(); err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uint
				_ = rows.Scan(&id)
				adminIDs = append(adminIDs, id)
			}
		} else {
			log.Printf("warning: gagal ambil admin ids: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)

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

		title := "Kelas Dihapus"
		body := fmt.Sprintf("Kelas %s telah dihapus dari sistem.", kelas.Nama)
		payload := map[string]interface{}{
			"type":       "delete_kelas",
			"kelas_id":   fmt.Sprintf("%d", kelas.ID),
			"kelas_nama": kelas.Nama,
		}

		if err := firebaseclient.NotifyUsers(context.Background(), "delete_kelas", title, body, payload, finalRecipients); err != nil {
			log.Printf("NotifyUsers error (delete_kelas): %v", err)
		}
	}(kelas, siswaIDs)

	utils.SuccessResponse(c, http.StatusOK, "Data kelas berhasil dihapus", nil)
}
