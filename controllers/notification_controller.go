package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func GetNotifications(c *gin.Context) {
	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
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

	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("per_page", "20")
	unreadOnly := c.DefaultQuery("unread_only", "false")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 || perPage > 200 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	db := database.DB.Model(&models.Notification{}).Where("recipient = ?", userID)
	if unreadOnly == "true" || unreadOnly == "1" {
		db = db.Where("`read` = ?", false)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung notifikasi: "+err.Error())
		return
	}

	var notifs []models.Notification
	if err := db.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&notifs).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil notifikasi: "+err.Error())
		return
	}

	items := make([]requests.NotificationItemResponse, 0, len(notifs))
	for _, n := range notifs {
		var payload map[string]interface{}
		if n.Payload != "" {
			if err := json.Unmarshal([]byte(n.Payload), &payload); err != nil {
				payload = map[string]interface{}{"_raw": n.Payload}
			}
		}
		items = append(items, requests.NotificationItemResponse{
			ID:        n.ID,
			Title:     n.Title,
			Body:      n.Body,
			Type:      n.Type,
			Payload:   payload,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		})
	}

	resp := gin.H{
		"data":        items,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
		"unread_only": unreadOnly == "true" || unreadOnly == "1",
	}

	utils.SuccessResponse(c, http.StatusOK, "Daftar notifikasi", resp)
}

func MarkNotificationRead(c *gin.Context) {
	idParam := c.Param("id")
	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "ID notifikasi tidak valid")
		return
	}
	id := uint(id64)

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
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

	var notif models.Notification
	if err := database.DB.First(&notif, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Notifikasi tidak ditemukan")
		return
	}
	if notif.Recipient != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki akses ke notifikasi ini")
		return
	}
	if notif.Read {
		utils.SuccessResponse(c, http.StatusOK, "Notifikasi sudah ditandai terbaca", nil)
		return
	}

	notif.Read = true
	if err := database.DB.Save(&notif).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menandai notifikasi terbaca: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notifikasi berhasil ditandai terbaca", gin.H{"id": notif.ID})
}

func MarkNotificationsRead(c *gin.Context) {
	var body struct {
		IDs []uint `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}
	if len(body.IDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "ids wajib diisi")
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
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

	if err := database.DB.Model(&models.Notification{}).
		Where("id IN ? AND recipient = ?", body.IDs, userID).
		Update("read", true).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menandai notifikasi terbaca: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notifikasi berhasil ditandai terbaca", gin.H{"updated_count": len(body.IDs)})
}

func MarkAllRead(c *gin.Context) {
	userIDVal, ok := c.Get("user_id")
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "akses ditolak: user tidak ditemukan di context")
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

	if err := database.DB.Model(&models.Notification{}).
		Where("recipient = ? AND `read` = ?", userID, false).
		Update("read", true).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menandai semua notifikasi terbaca: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Semua notifikasi berhasil ditandai terbaca", nil)
}
