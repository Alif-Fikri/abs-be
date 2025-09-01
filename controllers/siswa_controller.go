package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"abs-be/firebaseclient"
	"net/http"
	"strconv"
	"time"
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func CreateSiswa(c *gin.Context) {
	var req requests.CreateSiswaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var existingSiswa models.Siswa
	if err := database.DB.Where("nisn = ?", req.NISN).First(&existingSiswa).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "NISN sudah terdaftar")
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan format YYYY-MM-DD")
		return
	}

	plainPassword := tglLahir.Format("02012006")

	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi password")
		return
	}

	newSiswa := models.Siswa{
		Nama:         req.Nama,
		NISN:         req.NISN,
		TempatLahir:  req.TempatLahir,
		TanggalLahir: tglLahir,
		JenisKelamin: req.JenisKelamin,
		NamaAyah:     req.NamaAyah,
		NamaIbu:      req.NamaIbu,
		Alamat:       req.Alamat,
		Agama:        req.Agama,
		Email:        req.Email,
		Telepon:      req.Telepon,
		AsalSekolah:  req.AsalSekolah,
		Password:     hashedPassword,
	}

	// if req.KelasID != 0 { newSiswa.KelasID = req.KelasID }

	if err := database.DB.Create(&newSiswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat siswa: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Siswa berhasil dibuat", gin.H{
		"siswa": newSiswa,
	})

	go func(siswa models.Siswa) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic di notification goroutine CreateSiswa: %v", r)
			}
		}()

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
			log.Printf("CreateSiswa: gagal ambil admin ids: %v", err)
		}
		recipientIDs = append(recipientIDs, adminIDs...)

		if siswa.KelasID != nil && *siswa.KelasID != 0 {
			var kelas models.Kelas
			if err := database.DB.Preload("WaliKelas").First(&kelas, *siswa.KelasID).Error; err == nil {
				if kelas.WaliKelasID != nil && *kelas.WaliKelasID != 0 {
					recipientIDs = append(recipientIDs, *kelas.WaliKelasID)
				}
			}
		}

		recipientIDs = append(recipientIDs, siswa.ID)

		uniq := map[uint]struct{}{}
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

		title := "Akun Siswa Baru Dibuat"
		body := fmt.Sprintf("Akun siswa %s (%s) berhasil dibuat.", siswa.Nama, siswa.NISN)
		payload := map[string]interface{}{
			"type":     "create_siswa",
			"siswa_id": fmt.Sprintf("%d", siswa.ID),
			"nama":     siswa.Nama,
			"nisn":     siswa.NISN,
		}

		if err := firebaseclient.SendNotify(context.Background(), "create_siswa", title, body, payload, finalRecipients); err != nil {
			log.Printf("CreateSiswa: NotifyUsers failed: %v", err)
		}
	}(newSiswa)
}

func GetSiswaByKelas(c *gin.Context) {
	kelasIDParam := c.Param("id")
	kid, err := strconv.ParseUint(kelasIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelasId tidak valid")
		return
	}
	kidUint := uint(kid)

	db := database.DB

	var ids []uint
	if err := db.
		Table("siswas").
		Select("DISTINCT siswas.id").
		Joins("LEFT JOIN kelas_siswas ks ON ks.siswa_id = siswas.id").
		Where("siswas.kelas_id = ? OR ks.kelas_id = ?", kidUint, kidUint).
		Pluck("siswas.id", &ids).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil daftar siswa: "+err.Error())
		return
	}

	if len(ids) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Tidak ada siswa di kelas tersebut")
		return
	}

	var siswa []models.Siswa
	if err := db.Preload("Kelas").Preload("MataPelajaran").Find(&siswa, ids).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memuat relasi siswa: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa berhasil diambil", siswa)
}

func GetAllSiswa(c *gin.Context) {
	var siswa []models.Siswa
	if err := database.DB.Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data siswa")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil diambil", siswa)
}

func GetSiswaByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var siswa models.Siswa
	if err := database.DB.First(&siswa, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil ditemukan", siswa)
}

func UpdateSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var siswa models.Siswa
	if err := database.DB.First(&siswa, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	var req requests.CreateSiswaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan format YYYY-MM-DD")
		return
	}

	siswa.Nama = req.Nama
	siswa.NISN = req.NISN
	siswa.TempatLahir = req.TempatLahir
	siswa.TanggalLahir = tglLahir
	siswa.JenisKelamin = req.JenisKelamin
	siswa.NamaAyah = req.NamaAyah
	siswa.NamaIbu = req.NamaIbu
	siswa.Alamat = req.Alamat
	siswa.Agama = req.Agama
	siswa.Email = req.Email
	siswa.Telepon = req.Telepon
	siswa.AsalSekolah = req.AsalSekolah

	if err := database.DB.Save(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil diperbarui", siswa)
}

func DeleteSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	if err := database.DB.Delete(&models.Siswa{}, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus data siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Data siswa berhasil dihapus", nil)
}

func GetProfilSiswa(c *gin.Context) {
	siswaID := c.MustGet("user_id").(uint)

	var siswa models.Siswa
	if err := database.DB.First(&siswa, siswaID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}

	response := requests.SiswaResponse{
		ID:           siswa.ID,
		Nama:         siswa.Nama,
		NISN:         siswa.NISN,
		TempatLahir:  siswa.TempatLahir,
		TanggalLahir: siswa.TanggalLahir.Format("2006-01-02"),
		JenisKelamin: siswa.JenisKelamin,
		NamaAyah:     siswa.NamaAyah,
		NamaIbu:      siswa.NamaIbu,
		Alamat:       siswa.Alamat,
		Agama:        siswa.Agama,
		Email:        siswa.Email,
		Telepon:      siswa.Telepon,
		AsalSekolah:  siswa.AsalSekolah,
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil siswa", response)
}

func GetAbsensiSiswa(c *gin.Context) {
	siswaID := c.MustGet("user_id").(uint)

	tanggal := c.Query("tanggal")
	tipe := c.Query("tipe")

	db := database.DB

	var absensi []models.AbsensiSiswa
	q := db.
		Preload("Siswa").
		Preload("Kelas").
		Preload("MataPelajaran").
		Preload("Guru").
		Where("siswa_id = ?", siswaID)

	if tanggal != "" {
		if t, err := time.Parse("2006-01-02", tanggal); err == nil {
			q = q.Where("DATE(tanggal) = ?", t.Format("2006-01-02"))
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah, gunakan YYYY-MM-DD")
			return
		}
	}

	if tipe != "" {
		if tipe != "kelas" && tipe != "mapel" {
			utils.ErrorResponse(c, http.StatusBadRequest, "tipe harus 'kelas' atau 'mapel'")
			return
		}
		q = q.Where("tipe_absensi = ?", tipe)
	}

	if err := q.Order("tanggal ASC, id ASC").Find(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data absensi: "+err.Error())
		return
	}

	var resp []requests.AbsensiResponse
	for _, a := range absensi {

		siswaPub := requests.SiswaPublic{
			ID: a.SiswaID,
		}
		if a.Siswa.ID != 0 {
			siswaPub.Nama = a.Siswa.Nama
			siswaPub.NISN = a.Siswa.NISN
			siswaPub.JenisKelamin = a.Siswa.JenisKelamin
		}

		kelasPub := requests.KelasPublic{
			ID:          a.KelasID,
			Nama:        "",
			Tingkat:     "",
			TahunAjaran: a.TahunAjaran,
		}
		if a.Kelas.ID != 0 {
			kelasPub.Nama = a.Kelas.Nama
			kelasPub.Tingkat = a.Kelas.Tingkat
		}

		var mapelPub *requests.MapelPublic
		if a.MapelID != nil && a.MataPelajaran.ID != 0 {
			mapelPub = &requests.MapelPublic{
				ID:   a.MataPelajaran.ID,
				Nama: a.MataPelajaran.Nama,
				Kode: a.MataPelajaran.Kode,
			}
		}

		guruPub := requests.GuruPublic{
			ID: 0,
		}
		if a.Guru.ID != 0 {
			guruPub = requests.GuruPublic{
				ID:    a.Guru.ID,
				Nama:  a.Guru.Nama,
				NIP:   a.Guru.NIP,
				Email: a.Guru.Email,
			}
		} else if a.GuruID != 0 {
			guruPub.ID = a.GuruID
		}

		entry := requests.AbsensiResponse{
			ID:          a.ID,
			Siswa:       siswaPub,
			Kelas:       kelasPub,
			Mapel:       mapelPub,
			Guru:        guruPub,
			TipeAbsensi: a.TipeAbsensi,
			Tanggal:     a.Tanggal.Format("2006-01-02"),
			Status:      a.Status,
			Keterangan:  a.Keterangan,
			TahunAjaran: a.TahunAjaran,
			Semester:    a.Semester,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
		}
		resp = append(resp, entry)
	}

	utils.SuccessResponse(c, http.StatusOK, "Data absensi", resp)
}
