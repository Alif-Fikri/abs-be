package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"abs-be/database"
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
			utils.ErrorResponse(c, http.StatusForbidden, "Anda tidak mengajar mapel ini di kelas yang diminta")
			return
		}
	}

	guruIDForInsert := req.GuruID
	if role == "guru" {
		guruIDForInsert = uint(userID) 
	}

	var exist models.AbsensiSiswa
	q := database.DB.Where("siswa_id = ? AND tanggal = ? AND tipe_absensi = ? AND kelas_id = ?", req.SiswaID, tanggal, req.TipeAbsensi, req.KelasID)
	if req.MapelID != nil {
		q = q.Where("mapel_id = ?", req.MapelID)
	} else {
		q = q.Where("mapel_id IS NULL")
	}
	if err := q.First(&exist).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Absensi untuk siswa ini pada tanggal/tipe/mapel/kelas tersebut sudah ada")
		return
	}

	absensi := models.AbsensiSiswa{
		SiswaID:     req.SiswaID,
		KelasID:     req.KelasID,
		MapelID:     req.MapelID,
		GuruID:      uint(guruIDForInsert),
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

	utils.SuccessResponse(c, http.StatusCreated, "Absensi berhasil disimpan", absensi)
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui absensi")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Absensi berhasil diperbarui", absensi)
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

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk mapel", siswa)
}

func ListStudentsForKelas(c *gin.Context) {
	guruID := c.MustGet("user_id").(uint)

	var kelas models.Kelas
	if err := database.DB.Where("wali_kelas_id = ?", guruID).First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Anda belum ditetapkan wali kelas")
		return
	}

	var siswa []models.Siswa
	if err := database.DB.Where("kelas_id = ?", kelas.ID).Find(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil siswa")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar siswa untuk kelas Anda", siswa)
}

func RecapAbsensiMapel(c *gin.Context) {
	mapelID := c.Query("mapel_id")
	kelasID := c.Query("kelas_id")
	tgl := c.Query("tanggal")
	if mapelID == "" || kelasID == "" || tgl == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "mapel_id, kelas_id & tanggal wajib")
		return
	}
	tanggal, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	var recaps []struct {
		SiswaID uint
		Nama    string
		Status  string
	}
	if err := database.DB.
		Table("absensi_siswas").
		Select("siswa_id, siswas.nama, status").
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Where("tipe_absensi = ? AND mapel_id = ? AND kelas_id = ? AND tanggal = ?", "mapel", mapelID, kelasID, tanggal).
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
	tanggal, err := time.Parse("2006-01-02", tgl)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah")
		return
	}

	var recaps []struct {
		SiswaID uint
		Nama    string
		Status  string
	}
	if err := database.DB.
		Table("absensi_siswas").
		Select("siswa_id, siswas.nama, status").
		Joins("JOIN siswas ON siswas.id = absensi_siswas.siswa_id").
		Where("tipe_absensi = ? AND kelas_id = ? AND tanggal = ?", "kelas", kelasID, tanggal).
		Scan(&recaps).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal rekap absensi kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rekap absensi kelas", recaps)
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
	err := db.Table("guru_mapel_kelas").
		Select("mapel_id").
		Where("guru_id = ? AND kelas_id = ? AND tahun_ajaran = ? AND semester = ?", guruID, kelasID, ta, semester).
		Find(&rows).Error
	if err != nil {
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
	err := db.Table("guru_mapel_kelas").
		Select("mapel_id, kelas_id").
		Where("guru_id = ? AND tahun_ajaran = ? AND semester = ?", guruID, ta, semester).
		Find(&rows).Error
	if err != nil {
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
