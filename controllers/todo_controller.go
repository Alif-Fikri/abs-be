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

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	jamSekarang := time.Now().Format("15:04:05")

	todo := models.Todo{
		UserID:    userID,
		Role:      role,
		Tanggal:   req.Tanggal,
		Deskripsi: req.Deskripsi,
		JamDibuat: jamSekarang,
	}

	if err := database.DB.Create(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat catatan baru")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Catatan berhasil dibuat", todo)
}

func GetTodosByTanggal(c *gin.Context) {
	tanggal := c.Query("tanggal")
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02") 
	}

	userID := c.MustGet("userID").(uint)
	role := c.MustGet("role").(string)

	var todos []models.Todo
	if err := database.DB.Where("user_id = ? AND role = ? AND tanggal = ?", userID, role, tanggal).Order("jam_dibuat ASC").Find(&todos).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil catatan")
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

	var todo models.Todo
	if err := database.DB.First(&todo, todoID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Catatan tidak ditemukan")
		return
	}

	todo.IsDone = req.IsDone

	if err := database.DB.Save(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui status catatan")
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

	if err := database.DB.Delete(&models.Todo{}, todoID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus catatan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Catatan berhasil dihapus", nil)
}
