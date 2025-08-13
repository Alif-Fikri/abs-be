package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AssignMapelKelas(c *gin.Context) {
	var req requests.AssignMapelKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "panic terjadi")
		}
	}()

	var existsSameCombo models.GuruMapelKelas
	if err := tx.Where("kelas_id = ? AND mapel_id = ? AND tahun_ajaran = ? AND semester = ?",
		req.KelasID, req.MapelID, req.TahunAjaran, req.Semester).First(&existsSameCombo).Error; err == nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Kombinasi kelas+mapel pada periode ini sudah ditugaskan ke guru lain")
		return
	}

	var existing models.GuruMapelKelas
	if err := tx.Where("guru_id = ? AND mapel_id = ? AND kelas_id = ? AND tahun_ajaran = ? AND semester = ?",
		req.GuruID, req.MapelID, req.KelasID, req.TahunAjaran, req.Semester).First(&existing).Error; err == nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Guru sudah mengajar mapel ini di kelas dan periode tersebut")
		return
	}

	newAssign := models.GuruMapelKelas{
		GuruID:      req.GuruID,
		MapelID:     req.MapelID,
		KelasID:     req.KelasID,
		TahunAjaran: req.TahunAjaran,
		Semester:    req.Semester,
	}
	if err := tx.Create(&newAssign).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menetapkan guru ke mapel dan kelas: "+err.Error())
		return
	}

	var existingRole models.GuruRole
	if err := tx.Where("guru_id = ? AND role = 'guru_mapel' AND kelas_id = ? AND mapel_id = ?",
		req.GuruID, req.KelasID, req.MapelID).First(&existingRole).Error; err != nil {
		role := models.GuruRole{
			GuruID:  req.GuruID,
			Role:    "guru_mapel",
			KelasID: &req.KelasID,
			MapelID: &req.MapelID,
		}
		if err := tx.Create(&role).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat role guru_mapel: "+err.Error())
			return
		}
	}

	var full models.GuruMapelKelas
	if err := tx.Preload("Guru").Preload("MataPelajaran").Preload("Kelas").First(&full, newAssign.ID).Error; err != nil {

		if err2 := tx.Commit().Error; err2 != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err2.Error())
			return
		}
		utils.SuccessResponse(c, http.StatusCreated, "Guru berhasil ditetapkan ke mapel dan kelas", newAssign)
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Guru berhasil ditetapkan ke mapel dan kelas", full)
}

func UpdateAssignMapelKelas(c *gin.Context) {
	id := c.Param("id")

	var req requests.AssignMapelKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "panic terjadi")
		}
	}()

	var assignment models.GuruMapelKelas
	if err := tx.First(&assignment, id).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Data penugasan tidak ditemukan")
		return
	}

	var conflict models.GuruMapelKelas
	if err := tx.Where("kelas_id = ? AND mapel_id = ? AND tahun_ajaran = ? AND semester = ? AND id <> ?",
		req.KelasID, req.MapelID, req.TahunAjaran, req.Semester, assignment.ID).First(&conflict).Error; err == nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusConflict, "Kombinasi kelas+mapel+periode sudah ditugaskan ke guru lain")
		return
	}

	oldGuruID := assignment.GuruID
	oldMapelID := assignment.MapelID
	oldKelasID := assignment.KelasID

	assignment.GuruID = req.GuruID
	assignment.MapelID = req.MapelID
	assignment.KelasID = req.KelasID
	assignment.TahunAjaran = req.TahunAjaran
	assignment.Semester = req.Semester

	if err := tx.Save(&assignment).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui penugasan: "+err.Error())
		return
	}

	if err := tx.Where("guru_id = ? AND role = 'guru_mapel' AND kelas_id = ? AND mapel_id = ?",
		oldGuruID, oldKelasID, oldMapelID).Delete(&models.GuruRole{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus role lama: "+err.Error())
		return
	}

	var existingRole models.GuruRole
	if err := tx.Where("guru_id = ? AND role = 'guru_mapel' AND kelas_id = ? AND mapel_id = ?",
		req.GuruID, req.KelasID, req.MapelID).First(&existingRole).Error; err != nil {
		role := models.GuruRole{
			GuruID:  req.GuruID,
			Role:    "guru_mapel",
			KelasID: &req.KelasID,
			MapelID: &req.MapelID,
		}
		if err := tx.Create(&role).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat role guru_mapel: "+err.Error())
			return
		}
	}

	var full models.GuruMapelKelas
	if err := tx.Preload("Guru").Preload("MataPelajaran").Preload("Kelas").First(&full, assignment.ID).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data penugasan setelah update: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Penugasan berhasil diperbarui", full)
}

func DeleteAssignMapelKelas(c *gin.Context) {
	id := c.Param("id")

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "panic terjadi")
		}
	}()

	var assignment models.GuruMapelKelas
	if err := tx.First(&assignment, id).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusNotFound, "Data penugasan tidak ditemukan")
		return
	}

	if err := tx.Where("guru_id = ? AND role = 'guru_mapel' AND kelas_id = ? AND mapel_id = ?",
		assignment.GuruID, assignment.KelasID, assignment.MapelID).Delete(&models.GuruRole{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus role terkait: "+err.Error())
		return
	}

	if err := tx.Delete(&assignment).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus penugasan: "+err.Error())
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal commit transaksi: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Penugasan berhasil dihapus", nil)
}

func ListAssignMapelKelas(c *gin.Context) {
	var assignments []models.GuruMapelKelas
	if err := database.DB.
		Preload("Guru").
		Preload("MataPelajaran").
		Preload("Kelas").
		Find(&assignments).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data penugasan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar penugasan guru ke mapel dan kelas", assignments)
}
