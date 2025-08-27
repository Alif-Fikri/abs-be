package controllers

import (
	"net/http"
	"abs-be/database"
	"abs-be/models"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func RegisterDeviceToken(c *gin.Context) {
    userID := c.MustGet("user_id").(uint) 
    var req struct {
        Token    string `json:"token" binding:"required"`
        Platform string `json:"platform"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, "token wajib")
        return
    }

    dt := models.DeviceToken{
        UserID: userID,
        Token:  req.Token,
        Platform: req.Platform,
    }
    if err := database.DB.Where("token = ?", req.Token).Assign(dt).FirstOrCreate(&dt).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menyimpan device token")
        return
    }
    utils.SuccessResponse(c, http.StatusOK, "Device token terdaftar", nil)
}
