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

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)

	tanggal, err := time.Parse("2006-01-02", req.Tanggal)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan YYYY-MM-DD")
		return
	}

	todo := models.Todo{
		UserID:    userID,
		Role:      role,
		Tanggal:   tanggal,
		Deskripsi: req.Deskripsi,
		JamDibuat: time.Now().Format("15:04:05"),
	}

	if err := database.DB.Create(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat catatan baru: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Catatan berhasil dibuat", todo)
}

func GetTodosByTanggal(c *gin.Context) {
	tanggalStr := c.Query("tanggal")
	var tanggal time.Time
	var err error

	if tanggalStr == "" {
		tanggal = time.Now()
	} else {
		tanggal, err = time.Parse("2006-01-02", tanggalStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Format tanggal salah. Gunakan YYYY-MM-DD")
			return
		}
	}

	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)

	var todos []models.Todo
	if err := database.DB.Where("user_id = ? AND role = ? AND tanggal = ?", userID, role, tanggal.Format("2006-01-02")).Order("jam_dibuat ASC").Find(&todos).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil catatan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar catatan berhasil diambil", todos)
}

func UpdateTodoStatus(c *gin.Context) {
	todoIDParam := c.Param("id")
	todoID, err := strconv.Atoi(todoIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID catatan tidak valid")
		return
	}

	var req requests.UpdateTodoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	userID := c.MustGet("user_id").(uint)

	var todo models.Todo
	if err := database.DB.Where("id = ? AND user_id = ?", todoID, userID).First(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Catatan tidak ditemukan")
		return
	}

	todo.IsDone = req.IsDone

	if err := database.DB.Save(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui status catatan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Status catatan berhasil diperbarui", todo)
}

func DeleteTodo(c *gin.Context) {
	todoIDParam := c.Param("id")
	todoID, err := strconv.Atoi(todoIDParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID catatan tidak valid")
		return
	}

	userID := c.MustGet("user_id").(uint)

	if err := database.DB.Where("id = ? AND user_id = ?", todoID, userID).Delete(&models.Todo{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus catatan: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Catatan berhasil dihapus", nil)
}
