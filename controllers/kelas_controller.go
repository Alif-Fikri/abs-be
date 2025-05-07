package controllers

import (
	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateKelas(c *gin.Context) {
	var req requests.CreateKelasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
		return
	}

	newKelas := models.Kelas{
		Nama:        req.Nama,
		Tingkat:     req.Tingkat,
		TahunAjaran: req.TahunAjaran,
	}

	if err := database.DB.Create(&newKelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat kelas")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Kelas berhasil dibuat", newKelas)
}