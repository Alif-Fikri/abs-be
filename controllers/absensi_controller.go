package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"abs-be/database"
	"abs-be/firebaseclient"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateAbsensiSiswa(c *gin.Context) {
	var req requests.AbsensiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tanggal, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal harus YYYY-MM-DD")
		return
	}

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "role tidak ditemukan di context")
		return
	}
	role := roleVal.(string)

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user_id tidak ditemukan di context")
		return
	}
	var userID uint
	switch v := userIDVal.(type) {
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

	switch role {
	case "guru":
		if req.TipeAbsensi != "mapel" {
			utils.ErrorResponse(c, http.StatusForbidden, "Guru hanya dapat mengisi absen mapel")
			return
		}
	case "wali_kelas":
		if req.TipeAbsensi != "kelas" {
			utils.ErrorResponse(c, http.StatusForbidden, "Wali kelas hanya dapat mengisi absen kelas")
			return
		}
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Role tidak diizinkan")
		return
	}

	var s models.Siswa
	if err := database.DB.First(&s, req.SiswaID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Siswa tidak ditemukan")
		return
	}

	var count int64
	if err := database.DB.Table("kelas_siswas").
		Where("siswa_id = ? AND kelas_id = ?", req.SiswaID, req.KelasID).
		Count(&count).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa keanggotaan siswa: "+err.Error())
		return
	}
	if count == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Siswa tidak terdaftar di kelas yang diberikan")
		return
	}

	if req.TipeAbsensi == "mapel" {
		if req.MapelID == nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id harus diisi untuk absen mapel")
			return
		}

		ta := getTahunAjaranNow()
		sem := getSemesterNow()

		mapelIDs, err := getMapelIDsByGuruAndKelas(database.DB, userID, req.KelasID, ta, sem)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa pengajaran guru: "+err.Error())
			return
		}

		found := false
		for _, m := range mapelIDs {
			if m == *req.MapelID {
				found = true
				break
			}
		}
		if !found {
			mapelIDsLoose, err := getMapelIDsByGuruAndKelas(database.DB, userID, req.KelasID, "", "")
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa pengajaran guru (fallback): "+err.Error())
				return
			}
			for _, m := range mapelIDsLoose {
				if m == *req.MapelID {
					found = true
					break
				}
			}
		}
		if !found {
			utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar mapel ini di kelas yang diminta")
			return
		}
	}

	guruIDForInsert := req.GuruID
	if role == "guru" {
		guruIDForInsert = userID
	}

	var exist models.AbsensiSiswa
	q := database.DB.
		Where("siswa_id = ? AND DATE(tanggal) = ? AND tipe_absensi = ? AND kelas_id = ?",
			req.SiswaID, tanggal.Format("2006-01-02"), req.TipeAbsensi, req.KelasID)

	if req.MapelID != nil {
		q = q.Where("mapel_id = ?", req.MapelID)
	} else {
		q = q.Where("mapel_id IS NULL")
	}

	if role == "guru" {
		q = q.Where("guru_id = ?", userID)
	}

	if err := q.First(&exist).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Absensi untuk siswa ini pada tanggal/tipe/mapel/kelas oleh guru ini sudah ada")
		return
	}

	absensi := models.AbsensiSiswa{
		SiswaID:     req.SiswaID,
		KelasID:     req.KelasID,
		MapelID:     req.MapelID,
		GuruID:      guruIDForInsert,
		TipeAbsensi: req.TipeAbsensi,
		Tanggal:     tanggal,
		Status:      req.Status,
		Keterangan:  req.Keterangan,
		TahunAjaran: getTahunAjaranNow(),
		Semester:    getSemesterNow(),
	}

	if err := database.DB.Create(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan absensi: "+err.Error())
		return
	}

	var full models.AbsensiSiswa
	if err := database.DB.
		Preload("Siswa", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nisn", "tanggal_lahir", "jenis_kelamin")
		}).
		Preload("Kelas", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "tingkat", "tahun_ajaran")
		}).
		Preload("MataPelajaran", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "kode")
		}).
		Preload("Guru", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nip", "email")
		}).
		First(&full, absensi.ID).Error; err != nil {
		utils.SuccessResponse(c, http.StatusCreated, "Absensi berhasil disimpan", absensi)
		return
	}

	resp := requests.AbsensiResponse{
		ID: full.ID,
		Siswa: requests.SiswaPublic{
			ID:           full.Siswa.ID,
			Nama:         full.Siswa.Nama,
			NISN:         full.Siswa.NISN,
			JenisKelamin: full.Siswa.JenisKelamin,
		},
		Kelas: requests.KelasPublic{
			ID:          full.Kelas.ID,
			Nama:        full.Kelas.Nama,
			Tingkat:     full.Kelas.Tingkat,
			TahunAjaran: full.Kelas.TahunAjaran,
		},
		TipeAbsensi: full.TipeAbsensi,
		Tanggal:     full.Tanggal.Format("2006-01-02"),
		Status:      full.Status,
		Keterangan:  full.Keterangan,
		TahunAjaran: full.TahunAjaran,
		Semester:    full.Semester,
		CreatedAt:   full.CreatedAt,
		UpdatedAt:   full.UpdatedAt,
	}

	if full.MapelID != nil && full.MataPelajaran.ID != 0 {
		mp := requests.MapelPublic{
			ID:   full.MataPelajaran.ID,
			Nama: full.MataPelajaran.Nama,
			Kode: full.MataPelajaran.Kode,
		}
		resp.Mapel = &mp
	}

	resp.Guru = requests.GuruPublic{
		ID:    full.Guru.ID,
		Nama:  full.Guru.Nama,
		NIP:   full.Guru.NIP,
		Email: full.Guru.Email,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Absensi berhasil disimpan", resp)
}

func UpdateAbsensiSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}
	id := uint(id64)

	var req requests.AbsensiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	var absensi models.AbsensiSiswa
	if err := database.DB.First(&absensi, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Absensi tidak ditemukan")
		return
	}

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "format role tidak valid")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user_id tidak ditemukan di context")
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
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

	switch role {
	case "guru":
		if absensi.GuruID != userID {
			utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk mengedit absensi ini")
			return
		}
	case "wali_kelas":
		var kelas models.Kelas
		if err := database.DB.First(&kelas, absensi.KelasID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas terkait")
			return
		}
		if kelas.WaliKelasID == nil || *kelas.WaliKelasID != userID {
			utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk mengedit absensi ini")
			return
		}
	case "admin":
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Role tidak diizinkan")
		return
	}

	updated := false
	if req.Status != "" && req.Status != absensi.Status {
		absensi.Status = req.Status
		updated = true
	}
	if req.Keterangan != "" && req.Keterangan != absensi.Keterangan {
		absensi.Keterangan = req.Keterangan
		updated = true
	}

	if !updated {
		utils.ErrorResponse(c, http.StatusBadRequest, "Tidak ada perubahan yang valid untuk disimpan")
		return
	}

	if err := database.DB.Save(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui absensi: "+err.Error())
		return
	}

	var full models.AbsensiSiswa
	if err := database.DB.
		Preload("Siswa", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nisn", "tanggal_lahir", "jenis_kelamin")
		}).
		Preload("Kelas", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "tingkat", "tahun_ajaran")
		}).
		Preload("MataPelajaran", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "kode")
		}).
		Preload("Guru", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "nama", "nip", "email")
		}).
		First(&full, absensi.ID).Error; err != nil {
		respFallback := requests.AbsensiResponse{
			ID: absensi.ID,
			Siswa: requests.SiswaPublic{
				ID: absensi.SiswaID,
			},
			Kelas: requests.KelasPublic{
				ID: absensi.KelasID,
			},
			TipeAbsensi: absensi.TipeAbsensi,
			Tanggal:     absensi.Tanggal.Format("2006-01-02"),
			Status:      absensi.Status,
			Keterangan:  absensi.Keterangan,
			TahunAjaran: absensi.TahunAjaran,
			Semester:    absensi.Semester,
			CreatedAt:   absensi.CreatedAt,
			UpdatedAt:   absensi.UpdatedAt,
		}
		utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil diperbarui", respFallback)
		return
	}

	resp := requests.AbsensiResponse{
		ID: full.ID,
		Siswa: requests.SiswaPublic{
			ID:           full.Siswa.ID,
			Nama:         full.Siswa.Nama,
			NISN:         full.Siswa.NISN,
			JenisKelamin: full.Siswa.JenisKelamin,
		},
		Kelas: requests.KelasPublic{
			ID:          full.Kelas.ID,
			Nama:        full.Kelas.Nama,
			Tingkat:     full.Kelas.Tingkat,
			TahunAjaran: full.Kelas.TahunAjaran,
		},
		TipeAbsensi: full.TipeAbsensi,
		Tanggal:     full.Tanggal.Format("2006-01-02"),
		Status:      full.Status,
		Keterangan:  full.Keterangan,
		TahunAjaran: full.TahunAjaran,
		Semester:    full.Semester,
		CreatedAt:   full.CreatedAt,
		UpdatedAt:   full.UpdatedAt,
	}

	if full.MapelID != nil && full.MataPelajaran.ID != 0 {
		mp := requests.MapelPublic{
			ID:   full.MataPelajaran.ID,
			Nama: full.MataPelajaran.Nama,
			Kode: full.MataPelajaran.Kode,
		}
		resp.Mapel = &mp
	}

	resp.Guru = requests.GuruPublic{
		ID:    full.Guru.ID,
		Nama:  full.Guru.Nama,
		NIP:   full.Guru.NIP,
		Email: full.Guru.Email,
	}

	utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil diperbarui", resp)
}

func DeleteAbsensiSiswa(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}
	id := uint(id64)

	var absensi models.AbsensiSiswa
	if err := database.DB.First(&absensi, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Absensi tidak ditemukan")
		return
	}

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "format role tidak valid")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user_id tidak ditemukan di context")
		return
	}

	var userID uint
	switch v := userIDVal.(type) {
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

	switch role {
	case "guru":
		if absensi.GuruID != userID {
			utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menghapus absensi ini")
			return
		}
	case "wali_kelas":
		var kelas models.Kelas
		if err := database.DB.First(&kelas, absensi.KelasID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas terkait")
			return
		}
		if kelas.WaliKelasID == nil || *kelas.WaliKelasID != userID {
			utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menghapus absensi ini")
			return
		}
	case "admin":

	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Role tidak diizinkan")
		return
	}

	if err := database.DB.Delete(&absensi).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus absensi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil dihapus", nil)
}

func ListStudentsForMapel(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	if mapelID == "" || kelasID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id & kelas_id wajib")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelasID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil siswa")
		return
	}

	var siswaPublic []requests.SiswaPublic
	for _, s := range siswa {
		siswaPublic = append(siswaPublic, requests.SiswaPublic{
			ID:           s.ID,
			Nama:         s.Nama,
			NISN:         s.NISN,
			JenisKelamin: s.JenisKelamin,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk mapel", siswaPublic)
}

func ListStudentsForKelas(c *gin.Context) {
	kelasID := c.Query("kelas_id")
	if kelasID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id wajib")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelasID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil siswa")
		return
	}

	var siswaPublic []requests.SiswaPublic
	for _, s := range siswa {
		siswaPublic = append(siswaPublic, requests.SiswaPublic{
			ID:           s.ID,
			Nama:         s.Nama,
			NISN:         s.NISN,
			JenisKelamin: s.JenisKelamin,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk kelas", siswaPublic)
}

func RecapAbsensiMapel(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if mapelID == "" || kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id, kelas_id & tanggal wajib")
		return
	}

	_, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	type RecapAbsensiMapelResponse struct {
		SiswaID     uint   `json:"siswa_id"`
		NamaSiswa   string `json:"nama_siswa"`
		Status      string `json:"status"`
		Kelas       string `json:"kelas"`
		Mapel       string `json:"mapel"`
		Tanggal     string `json:"tanggal"`
		NamaGuru    string `json:"nama_guru"`
		TahunAjaran string `json:"tahun_ajaran"`
		Semester    string `json:"semester"`
	}

	var recaps []RecapAbsensiMapelResponse

	if err := database.DB.
		Table("absensi_siswas").
		Select(`
			absensi_siswas.siswa_id,
			siswas.nama AS nama_siswa,
			absensi_siswas.status,
			kelas.nama AS kelas,
			mata_pelajarans.nama AS mapel,
			absensi_siswas.tanggal AS tanggal,
			gurus.nama AS nama_guru,
			absensi_siswas.tahun_ajaran,
			absensi_siswas.semester
		`).
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Joins("JOIN kelas ON kelas.id = absensi_siswas.kelas_id").
		Joins("JOIN mata_pelajarans ON mata_pelajarans.id = absensi_siswas.mapel_id").
		Joins("JOIN gurus ON gurus.id = absensi_siswas.guru_id").
		Where("absensi_siswas.tipe_absensi = ? AND absensi_siswas.mapel_id = ? AND absensi_siswas.kelas_id = ? AND DATE(absensi_siswas.tanggal) = ?",
			"mapel", mapelID, kelasID, tgl).
		Scan(&recaps).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi mapel")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rekap absensi mapel", recaps)
}

func RecapAbsensiKelas(c *gin.Context) {
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id & tanggal wajib")
		return
	}

	_, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah (gunakan YYYY-MM-DD)")
		return
	}

	type RecapAbsensiKelasResponse struct {
		SiswaID     uint   `json:"siswa_id"`
		NamaSiswa   string `json:"nama_siswa"`
		Status      string `json:"status"`
		Kelas       string `json:"kelas"`
		Tanggal     string `json:"tanggal"`
		WaliKelas   string `json:"wali_kelas"`
		TahunAjaran string `json:"tahun_ajaran"`
		Semester    string `json:"semester"`
	}

	var recaps []RecapAbsensiKelasResponse

	if err := database.DB.
		Table("absensi_siswas").
		Select(`
			absensi_siswas.siswa_id,
			siswas.nama AS nama_siswa,
			absensi_siswas.status,
			kelas.nama AS kelas,
			absensi_siswas.tanggal AS tanggal,
			gurus.nama AS wali_kelas,
			absensi_siswas.tahun_ajaran,
			absensi_siswas.semester
		`).
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Joins("JOIN kelas ON kelas.id = absensi_siswas.kelas_id").
		Joins("JOIN gurus ON gurus.id = kelas.wali_kelas_id").
		Where("absensi_siswas.tipe_absensi = ? AND absensi_siswas.kelas_id = ? AND DATE(absensi_siswas.tanggal) = ?",
			"kelas", kelasID, tgl).
		Scan(&recaps).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rekap absensi kelas", recaps)
}

func ExportRecapAbsensiMapelCSV(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if mapelID == "" || kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id, kelas_id & tanggal wajib")
		return
	}

	if _, err := time.Parse("2006-01-02", tgl); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah (gunakan YYYY-MM-DD)")
		return
	}

	userIDVal, ok := c.Get("user_id")
	var requesterID uint
	if ok {
		switch v := userIDVal.(type) {
		case uint:
			requesterID = v
		case int:
			requesterID = uint(v)
		case int64:
			requesterID = uint(v)
		case float64:
			requesterID = uint(v)
		default:
			requesterID = 0
		}
	} else {
		requesterID = 0
	}

	mid64, _ := strconv.ParseUint(mapelID, 10, 64)
	kid64, _ := strconv.ParseUint(kelasID, 10, 64)
	mid := uint(mid64)
	kid := uint(kid64)

	type row struct {
		NamaSiswa   string    `json:"nama_siswa"`
		Status      string    `json:"status"`
		Kelas       string    `json:"kelas"`
		Mapel       string    `json:"mapel"`
		NamaGuru    string    `json:"nama_guru"`
		TahunAjaran string    `json:"tahun_ajaran"`
		Semester    string    `json:"semester"`
		Tanggal     time.Time `json:"tanggal"`
	}

	var rows []row
	db := database.DB

	if err := db.
		Table("absensi_siswas").
		Select(`siswas.nama AS nama_siswa,
                absensi_siswas.status,
                kelas.nama AS kelas,
                mata_pelajarans.nama AS mapel,
                gurus.nama AS nama_guru,
                absensi_siswas.tahun_ajaran,
                absensi_siswas.semester,
                absensi_siswas.tanggal`).
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Joins("JOIN kelas ON kelas.id = absensi_siswas.kelas_id").
		Joins("JOIN mata_pelajarans ON mata_pelajarans.id = absensi_siswas.mapel_id").
		Joins("JOIN gurus ON gurus.id = absensi_siswas.guru_id").
		Where("absensi_siswas.tipe_absensi = ? AND absensi_siswas.mapel_id = ? AND absensi_siswas.kelas_id = ? AND DATE(absensi_siswas.tanggal) = ?",
			"mapel", mapelID, kelasID, tgl).
		Order("siswas.nama ASC").
		Scan(&rows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi mapel: "+err.Error())
		return
	}

	filename := fmt.Sprintf("rekap_mapel_%s_kelas_%s_%s.csv", mapelID, kelasID, tgl)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	if err := w.Write([]string{
		"Nama Siswa", "Status", "Kelas", "Mapel", "Guru", "Tahun Ajaran", "Semester", "Tanggal",
	}); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat CSV")
		return
	}

	for _, r := range rows {
		tStr := r.Tanggal.In(time.Local).Format("2006-01-02")
		record := []string{
			r.NamaSiswa,
			r.Status,
			r.Kelas,
			r.Mapel,
			r.NamaGuru,
			r.TahunAjaran,
			r.Semester,
			tStr,
		}
		if err := w.Write(record); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menulis CSV: "+err.Error())
			return
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyelesaikan CSV: "+err.Error())
		return
	}

	if requesterID != 0 {
		go func(reqID uint, mid, kid uint, tanggal, filename string, rowsCount int) {
			title := "Export Rekap Mapel Selesai"
			body := fmt.Sprintf("Rekap absensi mapel untuk tanggal %s telah selesai (%s).", tanggal, filename)
			payload := map[string]interface{}{
				"type":         "export_rekap_mapel",
				"mapel_id":     fmt.Sprintf("%d", mid),
				"kelas_id":     fmt.Sprintf("%d", kid),
				"tanggal":      tanggal,
				"filename":     filename,
				"record_count": fmt.Sprintf("%d", rowsCount),
			}
			if err := firebaseclient.SendNotify(context.Background(), "export_rekap_mapel", title, body, payload, []uint{reqID}); err != nil {
				log.Printf("ExportMapel: SendNotify error for user %d: %v", reqID, err)
			}
		}(requesterID, mid, kid, tgl, filename, len(rows))
	} else {
		log.Printf("ExportMapel: user_id not found in context, skipping personal notification")
	}
}

func ExportRecapAbsensiKelasCSV(c *gin.Context) {
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id & tanggal wajib")
		return
	}

	if _, err := time.Parse("2006-01-02", tgl); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah (gunakan YYYY-MM-DD)")
		return
	}

	userIDVal, ok := c.Get("user_id")
	var requesterID uint
	if ok {
		switch v := userIDVal.(type) {
		case uint:
			requesterID = v
		case int:
			requesterID = uint(v)
		case int64:
			requesterID = uint(v)
		case float64:
			requesterID = uint(v)
		default:
			requesterID = 0
		}
	} else {
		requesterID = 0
	}

	kid64, _ := strconv.ParseUint(kelasID, 10, 64)
	kid := uint(kid64)

	type row struct {
		NamaSiswa   string    `json:"nama_siswa"`
		Status      string    `json:"status"`
		Kelas       string    `json:"kelas"`
		WaliKelas   string    `json:"wali_kelas"`
		TahunAjaran string    `json:"tahun_ajaran"`
		Semester    string    `json:"semester"`
		Tanggal     time.Time `json:"tanggal"`
	}

	var rows []row
	db := database.DB

	if err := db.
		Table("absensi_siswas").
		Select(`siswas.nama AS nama_siswa,
                absensi_siswas.status,
                kelas.nama AS kelas,
                gurus.nama AS wali_kelas,
                absensi_siswas.tahun_ajaran,
                absensi_siswas.semester,
                absensi_siswas.tanggal`).
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Joins("JOIN kelas ON kelas.id = absensi_siswas.kelas_id").
		Joins("JOIN gurus ON gurus.id = kelas.wali_kelas_id").
		Where("absensi_siswas.tipe_absensi = ? AND absensi_siswas.kelas_id = ? AND DATE(absensi_siswas.tanggal) = ?",
			"kelas", kelasID, tgl).
		Order("siswas.nama ASC").
		Scan(&rows).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi kelas: "+err.Error())
		return
	}

	filename := fmt.Sprintf("rekap_kelas_%s_%s.csv", kelasID, tgl)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Write([]byte("\xEF\xBB\xBF"))

	w := csv.NewWriter(c.Writer)
	defer w.Flush()
	if err := w.Write([]string{
		"Nama Siswa", "Status", "Kelas", "Wali Kelas", "Tahun Ajaran", "Semester", "Tanggal",
	}); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat CSV")
		return
	}

	for _, r := range rows {
		tStr := r.Tanggal.In(time.Local).Format("2006-01-02")
		record := []string{
			r.NamaSiswa,
			r.Status,
			r.Kelas,
			r.WaliKelas,
			r.TahunAjaran,
			r.Semester,
			tStr,
		}
		if err := w.Write(record); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menulis CSV: "+err.Error())
			return
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyelesaikan CSV: "+err.Error())
		return
	}

	if requesterID != 0 {
		go func(reqID uint, kid uint, tanggal, filename string, rowsCount int) {
			title := "Export Rekap Kelas Selesai"
			body := fmt.Sprintf("Rekap absensi kelas untuk tanggal %s telah selesai (%s).", tanggal, filename)
			payload := map[string]interface{}{
				"type":         "export_rekap_kelas",
				"kelas_id":     fmt.Sprintf("%d", kid),
				"tanggal":      tanggal,
				"filename":     filename,
				"record_count": fmt.Sprintf("%d", rowsCount),
			}
			if err := firebaseclient.SendNotify(context.Background(), "export_rekap_kelas", title, body, payload, []uint{reqID}); err != nil {
				log.Printf("ExportKelas: SendNotify error for user %d: %v", reqID, err)
			}
		}(requesterID, kid, tgl, filename, len(rows))
	} else {
		log.Printf("ExportKelas: user_id not found in context, skipping personal notification")
	}
}

func getTahunAjaranNow() string {
	now := time.Now()
	year := now.Year()
	if now.Month() >= 7 {
		return fmt.Sprintf("%d/%d", year, year+1)
	}
	return fmt.Sprintf("%d/%d", year-1, year)
}

func getSemesterNow() string {
	m := time.Now().Month()
	if m >= 7 && m <= 12 {
		return "ganjil"
	}
	return "genap"
}

func GetAbsensi(c *gin.Context) {
	kelasIDStr := c.Query("kelas_id")
	mapelIDStr := c.Query("mapel_id")
	tanggalStr := c.Query("tanggal")
	siswaIDStr := c.Query("siswa_id")

	var tanggal time.Time
	var err error
	if tanggalStr == "" {
		tanggal = time.Now()
	} else {
		tanggal, err = time.Parse("2006-01-02", tanggalStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "format tanggal salah, gunakan YYYY-MM-DD")
			return
		}
	}
	dateStr := tanggal.Format("2006-01-02")

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "format role tidak valid")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user_id tidak ditemukan di context")
		return
	}
	var userID uint
	switch v := userIDVal.(type) {
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

	db := database.DB

	where := db.Table("absensi_siswas").
		Select("absensi_siswas.id, absensi_siswas.siswa_id, siswas.nama as nama_siswa, absensi_siswas.kelas_id, absensi_siswas.mapel_id, mata_pelajarans.nama as nama_mapel, absensi_siswas.guru_id, absensi_siswas.tipe_absensi, DATE_FORMAT(absensi_siswas.tanggal, '%Y-%m-%d') as tanggal, absensi_siswas.status, absensi_siswas.keterangan, absensi_siswas.tahun_ajaran, absensi_siswas.semester").
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Joins("LEFT JOIN mata_pelajarans ON mata_pelajarans.id = absensi_siswas.mapel_id")

	where = where.Where("absensi_siswas.tanggal = ?", dateStr)

	if siswaIDStr != "" {
		if sid, err := strconv.ParseUint(siswaIDStr, 10, 64); err == nil {
			where = where.Where("absensi_siswas.siswa_id = ?", uint(sid))
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "siswa_id tidak valid")
			return
		}
	}

	switch role {
	case "admin":
		if kelasIDStr != "" {
			if kid, err := strconv.ParseUint(kelasIDStr, 10, 64); err == nil {
				where = where.Where("absensi_siswas.kelas_id = ?", uint(kid))
			} else {
				utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id tidak valid")
				return
			}
		}
		if mapelIDStr != "" {
			if mid, err := strconv.ParseUint(mapelIDStr, 10, 64); err == nil {
				where = where.Where("absensi_siswas.mapel_id = ?", uint(mid))
			} else {
				utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id tidak valid")
				return
			}
		}
	case "wali_kelas":
		var kelas models.Kelas
		if kelasIDStr == "" {
			if err := db.Where("wali_kelas_id = ?", userID).First(&kelas).Error; err != nil {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda belum ditetapkan sebagai wali kelas (tidak ada kelas ditemukan untuk user ini)")
				return
			}
			where = where.Where("absensi_siswas.kelas_id = ?", kelas.ID)
		} else {
			kid, err := strconv.ParseUint(kelasIDStr, 10, 64)
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id tidak valid")
				return
			}
			if err := db.First(&kelas, uint(kid)).Error; err != nil {
				utils.ErrorResponse(c, http.StatusNotFound, "kelas tidak ditemukan")
				return
			}
			if kelas.WaliKelasID == nil || *kelas.WaliKelasID != userID {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda bukan wali kelas untuk kelas yang diminta")
				return
			}
			where = where.Where("absensi_siswas.kelas_id = ?", uint(kid))
		}
		if mapelIDStr != "" {
			mid, err := strconv.ParseUint(mapelIDStr, 10, 64)
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id tidak valid")
				return
			}
			where = where.Where("absensi_siswas.mapel_id = ?", uint(mid))
		}
	case "guru":
		curTA := getTahunAjaranNow()
		curSem := getSemesterNow()

		if kelasIDStr != "" {
			kid, err := strconv.ParseUint(kelasIDStr, 10, 64)
			if err != nil {
				utils.ErrorResponse(c, http.StatusBadRequest, "kelas_id tidak valid")
				return
			}
			mapelIDs, err := getMapelIDsByGuruAndKelas(db, userID, uint(kid), curTA, curSem)
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "gagal memeriksa data pengajaran guru: "+err.Error())
				return
			}
			if len(mapelIDs) == 0 {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar di kelas yang diminta")
				return
			}
			if mapelIDStr != "" {
				mid, err := strconv.ParseUint(mapelIDStr, 10, 64)
				if err != nil {
					utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id tidak valid")
					return
				}
				found := false
				for _, m := range mapelIDs {
					if m == uint(mid) {
						found = true
						break
					}
				}
				if !found {
					utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar mapel ini di kelas yang diminta")
					return
				}
				where = where.Where("absensi_siswas.kelas_id = ? AND absensi_siswas.mapel_id = ?", uint(kid), uint(mid))
			} else {
				where = where.Where("absensi_siswas.kelas_id = ? AND absensi_siswas.mapel_id IN ?", uint(kid), mapelIDs)
			}
		} else {
			mapelIDsAll, kelasIDsAll, err := getMapelAndKelasByGuru(db, userID, curTA, curSem)
			if err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "gagal memeriksa data pengajaran guru: "+err.Error())
				return
			}
			if len(kelasIDsAll) == 0 {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda belum ditetapkan mengajar di kelas apapun")
				return
			}
			if mapelIDStr != "" {
				mid, err := strconv.ParseUint(mapelIDStr, 10, 64)
				if err != nil {
					utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id tidak valid")
					return
				}
				found := false
				for _, m := range mapelIDsAll {
					if m == uint(mid) {
						found = true
						break
					}
				}
				if !found {
					utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar mapel ini")
					return
				}
				where = where.Where("absensi_siswas.mapel_id = ? AND absensi_siswas.kelas_id IN ?", uint(mid), kelasIDsAll)
			} else {
				where = where.Where("absensi_siswas.kelas_id IN ? AND absensi_siswas.mapel_id IN ?", kelasIDsAll, mapelIDsAll)
			}
		}
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "role tidak diizinkan")
		return
	}

	var results []models.AbsensiResult
	if err := where.Order("absensi_siswas.kelas_id ASC, absensi_siswas.mapel_id ASC, absensi_siswas.siswa_id ASC").
		Scan(&results).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data absensi: "+err.Error())
		return
	}

	if len(results) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Data absensi tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar absensi", results)
}

func getMapelIDsByGuruAndKelas(db *gorm.DB, guruID uint, kelasID uint, ta string, semester string) ([]uint, error) {
	var rows []struct {
		MapelID uint `gorm:"column:mapel_id"`
	}

	q := db.Table("guru_mapel_kelas").Select("mapel_id").
		Where("guru_id = ? AND kelas_id = ?", guruID, kelasID)

	if ta != "" {
		q = q.Where("tahun_ajaran = ?", ta)
	}
	if semester != "" {
		q = q.Where("semester = ?", semester)
	}

	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.MapelID)
	}
	return out, nil
}

func getMapelAndKelasByGuru(db *gorm.DB, guruID uint, ta string, semester string) ([]uint, []uint, error) {
	var rows []struct {
		MapelID uint `gorm:"column:mapel_id"`
		KelasID uint `gorm:"column:kelas_id"`
	}
	q := db.Table("guru_mapel_kelas").Select("mapel_id, kelas_id").Where("guru_id = ?", guruID)
	if ta != "" {
		q = q.Where("tahun_ajaran = ?", ta)
	}
	if semester != "" {
		q = q.Where("semester = ?", semester)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	mapelSet := map[uint]struct{}{}
	kelasSet := map[uint]struct{}{}
	for _, r := range rows {
		mapelSet[r.MapelID] = struct{}{}
		kelasSet[r.KelasID] = struct{}{}
	}
	mapelIDs := make([]uint, 0, len(mapelSet))
	kelasIDs := make([]uint, 0, len(kelasSet))
	for k := range mapelSet {
		mapelIDs = append(mapelIDs, k)
	}
	for k := range kelasSet {
		kelasIDs = append(kelasIDs, k)
	}
	return mapelIDs, kelasIDs, nil
}

func GetAbsensiByID(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}
	id := uint(id64)

	var absensi models.AbsensiSiswa
	if err := database.DB.First(&absensi, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Absensi tidak ditemukan")
		return
	}

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "role tidak ditemukan di context")
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "format role tidak valid")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user_id tidak ditemukan di context")
		return
	}
	var userID uint
	switch v := userIDVal.(type) {
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

	switch role {
	case "admin":
	case "wali_kelas":
		var kelas models.Kelas
		if err := database.DB.First(&kelas, absensi.KelasID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data kelas")
			return
		}
		if kelas.WaliKelasID == nil || *kelas.WaliKelasID != userID {
			utils.ErrorResponse(c, http.StatusForbidden, "Anda bukan wali kelas untuk absensi ini")
			return
		}
	case "guru":
		curTA := getTahunAjaranNow()
		curSem := getSemesterNow()
		mapelIDs, err := getMapelIDsByGuruAndKelas(database.DB, userID, absensi.KelasID, curTA, curSem)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memeriksa hubungan guru/mapel: "+err.Error())
			return
		}

		if absensi.MapelID != nil {
			found := false
			for _, mid := range mapelIDs {
				if mid == *absensi.MapelID {
					found = true
					break
				}
			}
			if !found {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar mapel ini di kelas terkait")
				return
			}
		} else {
			if len(mapelIDs) == 0 {
				utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar di kelas ini")
				return
			}
		}
	default:
		utils.ErrorResponse(c, http.StatusForbidden, "Role tidak diizinkan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Detail absensi", absensi)
}
