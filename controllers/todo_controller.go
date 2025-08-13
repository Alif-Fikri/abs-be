package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateTodo(c *gin.Context) {
	var req requests.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
		return
	}
	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "user_id format tidak dikenali")
		return
	}

	role := roleVal.(string)

	tgl, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan YYYY-MM-DD")
		return
	}

	jamDibuat := time.Now().Format("15:04:05")

	todo := models.Todo{
		Role:      role,
		Tanggal:   tgl,
		Deskripsi: req.Deskripsi,
		JamDibuat: jamDibuat,
		IsDone:    false,
	}

	switch role {
	case "admin":
		todo.AdminID = &userID
	case "guru":
		todo.GuruID = &userID
	case "wali_kelas":
		todo.WaliKelasID = &userID
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak valid")
		return
	}

	if err := database.DB.Create(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat catatan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Catatan berhasil dibuat", todo)
}

func GetTodosByTanggal(c *gin.Context) {
	tanggalStr := c.Query("tanggal")
	var tgl time.Time
	var err error

	if tanggalStr == "" {
		tgl = time.Now()
	} else {
		tgl, err = time.Parse("2006-01-02", tanggalStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan YYYY-MM-DD")
			return
		}
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: user tidak ditemukan di context")
		return
	}

	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Akses ditolak: role tidak ditemukan di context")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "user_id format tidak dikenali")
		return
	}

	role, ok := roleVal.(string)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "role format tidak dikenali")
		return
	}

	var todos []models.Todo
	dateStr := tgl.Format("2006-01-02")
	db := database.DB

	switch role {
	case "admin":
		db = db.Where("admin_id = ? AND role = ? AND tanggal = ?", userID, "admin", dateStr)
	case "guru":
		db = db.Where("guru_id = ? AND role = ? AND tanggal = ?", userID, "guru", dateStr)
	case "wali_kelas":
		db = db.Where("wali_kelas_id = ? AND role = ? AND tanggal = ?", userID, "wali_kelas", dateStr)
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak valid")
		return
	}

	if err := db.Order("jam_dibuat ASC").Find(&todos).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil catatan: "+err.Error())
		return
	}

	if len(todos) == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Tidak ada catatan pada tanggal "+dateStr)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Daftar catatan berhasil diambil", todos)
}

func UpdateTodoStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var req requests.UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
		return
	}
	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "user_id format tidak dikenali")
		return
	}
	role := roleVal.(string)

	var todo models.Todo
	db := database.DB
	var findErr error

	switch role {
	case "admin":
		findErr = db.Where("id = ? AND admin_id = ?", id, userID).First(&todo).Error
	case "guru":
		findErr = db.Where("id = ? AND guru_id = ?", id, userID).First(&todo).Error
	case "wali_kelas":
		findErr = db.Where("id = ? AND wali_kelas_id = ?", id, userID).First(&todo).Error
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak valid")
		return
	}

	if findErr != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Catatan tidak ditemukan atau akses ditolak")
		return
	}

	todo.IsDone = req.IsDone
	if err := db.Save(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui status: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Status catatan berhasil diperbarui", todo)
}

func DeleteTodo(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID tidak valid")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
		return
	}
	roleVal, ok := c.Get("role")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: role tidak ditemukan di context")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "user_id format tidak dikenali")
		return
	}
	role := roleVal.(string)

	db := database.DB
	var delErr error

	switch role {
	case "admin":
		delErr = db.Where("id = ? AND admin_id = ?", id, userID).Delete(&models.Todo{}).Error
	case "guru":
		delErr = db.Where("id = ? AND guru_id = ?", id, userID).Delete(&models.Todo{}).Error
	case "wali_kelas":
		delErr = db.Where("id = ? AND wali_kelas_id = ?", id, userID).Delete(&models.Todo{}).Error
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak valid")
		return
	}

	if delErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus catatan: "+delErr.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Catatan berhasil dihapus", nil)
}
