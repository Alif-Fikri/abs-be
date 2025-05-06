package controllers

import (
	"net/http"

	"abs-be/database"
	"abs-be/models"
	"abs-be/requests"
	"abs-be/utils"

	"github.com/gin-gonic/gin"
)

func LoginAdmin(c *gin.Context) {
	var req requests.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	var admin models.Admin
	if err := database.DB.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan")
		return
	}

	if !utils.CheckPasswordHash(req.Password, admin.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
		return
	}

	token, err := utils.GenerateToken(admin.ID, "admin")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}

	session := models.Session{
		UserID: admin.ID,
		Token:  token,
		Role:   "admin",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"admin": requests.LoginResponse{ID: admin.ID, Email: admin.Email, Nama: admin.Nama},
	})
}

func LoginGuru(c *gin.Context) {
	var req requests.GuruLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	var guru models.Guru
	if err := database.DB.Where("email = ?", req.Email).First(&guru).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan")
		return
	}

	if !utils.CheckPasswordHash(req.Password, guru.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
		return
	}

	token, err := utils.GenerateToken(guru.ID, "guru")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}

	session := models.Session{
		UserID: guru.ID,
		Token:  token,
		Role:   "guru",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"guru":  requests.LoginResponse{ID: guru.ID, Email: guru.Email, Nama: guru.Nama},
	})
}

func LoginWaliKelas(c *gin.Context) {
	var req requests.GuruLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	var wali models.Guru
	if err := database.DB.
		Joins("JOIN guru_roles ON guru_roles.guru_id = gurus.id").
		Where("gurus.email = ? AND guru_roles.role = ?", req.Email, "wali_kelas").
		First(&wali).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan atau bukan wali kelas")
		return
	}

	if !utils.CheckPasswordHash(req.Password, wali.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
		return
	}

	token, err := utils.GenerateToken(wali.ID, "wali_kelas")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}

	session := models.Session{
		UserID: wali.ID,
		Token:  token,
		Role:   "wali_kelas",
	}

	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"wali_kelas": requests.LoginResponse{
			ID:    wali.ID,
			Email: wali.Email,
			Nama:  wali.Nama,
		},
	})
}

func Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "token tidak ditemukan")
		return
	}

	if err := database.DB.Where("token = ?", token).Delete(&models.Session{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal logout")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "logout berhasil", nil)
}
