package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateMataPelajaran(c *gin.Context) {
	var req requests.CreateMapelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	newMapel := models.MataPelajaran{
		Nama:    req.Nama,
		Kode:    req.Kode,
		Tingkat: req.Tingkat,
	}

	if err := database.DB.Create(&newMapel).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat mata pelajaran")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Mata pelajaran berhasil dibuat", newMapel)
}