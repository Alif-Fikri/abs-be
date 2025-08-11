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

	var waliKelasCount int64
	database.DB.Model(&models.GuruRole{}).
		Where("guru_id = ? AND role = 'wali_kelas'", guru.ID).
		Count(&waliKelasCount)
	isWaliKelas := waliKelasCount > 0

	if isWaliKelas {
		utils.ErrorResponse(c, http.StatusForbidden,
			"Anda adalah wali kelas, silakan login melalui endpoint wali kelas")
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

	var kelas models.Kelas
	if err := database.DB.
		Where("wali_kelas_id = ?", wali.ID).
		First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized,
			"wali kelas tidak memiliki kelas yang ditugaskan")
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
		"kelas": gin.H{
			"id":      kelas.ID,
			"nama":    kelas.Nama,
			"tingkat": kelas.Tingkat,
		},
	})
}

// func Logout(c *gin.Context) {
// 	token := c.GetHeader("Authorization")
// 	if token == "" {
// 		utils.ErrorResponse(c, http.StatusBadRequest, "token tidak ditemukan")
// 		return
// 	}

// 	if err := database.DB.Where("token = ?", token).Delete(&models.Session{}).Error; err != nil {
// 		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal logout")
// 		return
// 	}

// 	utils.SuccessResponse(c, http.StatusOK, "logout berhasil", nil)
// }

func LoginAll(c *gin.Context) {
	var req requests.AllLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	switch req.Role {
	case "admin":
		loginAdminAll(c, req)
	case "guru":
		loginGuruAll(c, req)
	case "wali_kelas":
		loginWaliKelasAll(c, req)
	case "siswa":
		loginSiswaAll(c, req)
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "role tidak dikenali")
	}
}

func loginAdminAll(c *gin.Context, req requests.AllLoginRequest) {
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
	session := models.Session{UserID: admin.ID, Token: token, Role: "admin"}
	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  requests.LoginResponse{ID: admin.ID, Email: admin.Email, Nama: admin.Nama},
		"role":  "admin",
	})
}

func loginGuruAll(c *gin.Context, req requests.AllLoginRequest) {
	var guru models.Guru
	if err := database.DB.Where("email = ?", req.Email).First(&guru).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan")
		return
	}
	if !utils.CheckPasswordHash(req.Password, guru.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
		return
	}
	var waliKelasCount int64
	database.DB.Model(&models.GuruRole{}).
		Where("guru_id = ? AND role = 'wali_kelas'", guru.ID).
		Count(&waliKelasCount)
	if waliKelasCount > 0 {
		utils.ErrorResponse(c, http.StatusForbidden, "Anda adalah wali kelas, silakan login sebagai wali kelas")
		return
	}
	token, err := utils.GenerateToken(guru.ID, "guru")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}
	session := models.Session{UserID: guru.ID, Token: token, Role: "guru"}
	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  requests.LoginResponse{ID: guru.ID, Email: guru.Email, Nama: guru.Nama},
		"role":  "guru",
	})
}

func loginWaliKelasAll(c *gin.Context, req requests.AllLoginRequest) {
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
	var kelas models.Kelas
	if err := database.DB.
		Where("wali_kelas_id = ?", wali.ID).
		First(&kelas).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "wali kelas belum ditugaskan ke kelas manapun")
		return
	}
	token, err := utils.GenerateToken(wali.ID, "wali_kelas")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}
	session := models.Session{UserID: wali.ID, Token: token, Role: "wali_kelas"}
	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  requests.LoginResponse{ID: wali.ID, Email: wali.Email, Nama: wali.Nama},
		"role":  "wali_kelas",
		"kelas": gin.H{
			"id":      kelas.ID,
			"nama":    kelas.Nama,
			"tingkat": kelas.Tingkat,
		},
	})
}

func loginSiswaAll(c *gin.Context, req requests.AllLoginRequest) {
	var siswa models.Siswa
	if err := database.DB.Where("email = ?", req.Email).First(&siswa).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan")
		return
	}
	if !utils.CheckPasswordHash(req.Password, siswa.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
		return
	}
	token, err := utils.GenerateToken(siswa.ID, "siswa")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token")
		return
	}
	session := models.Session{UserID: siswa.ID, Token: token, Role: "siswa"}
	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi login")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
		"token": token,
		"user":  requests.LoginResponse{ID: siswa.ID, Email: siswa.Email, Nama: siswa.Nama},
		"role":  "siswa",
	})
}

func LoginAutoRole(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	var admin models.Admin
	if err := database.DB.Where("email = ?", req.Email).First(&admin).Error; err == nil {
		if !utils.CheckPasswordHash(req.Password, admin.Password) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
			return
		}
		token, _ := utils.GenerateToken(admin.ID, "admin")
		database.DB.Create(&models.Session{UserID: admin.ID, Token: token, Role: "admin"})
		utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
			"token": token,
			"role":  "admin",
		})
		return
	}

	var guru models.Guru
	if err := database.DB.Where("email = ?", req.Email).First(&guru).Error; err == nil {
		if !utils.CheckPasswordHash(req.Password, guru.Password) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
			return
		}
		var waliKelasCount int64
		database.DB.Model(&models.GuruRole{}).
			Where("guru_id = ? AND role = 'wali_kelas'", guru.ID).
			Count(&waliKelasCount)
		role := "guru"
		if waliKelasCount > 0 {
			role = "wali_kelas"
		}
		token, _ := utils.GenerateToken(guru.ID, role)
		database.DB.Create(&models.Session{UserID: guru.ID, Token: token, Role: role})
		utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
			"token": token,
			"role":  role,
		})
		return
	}

	var siswa models.Siswa
	if err := database.DB.Where("email = ?", req.Email).First(&siswa).Error; err == nil {
		if !utils.CheckPasswordHash(req.Password, siswa.Password) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "password salah")
			return
		}
		token, _ := utils.GenerateToken(siswa.ID, "siswa")
		database.DB.Create(&models.Session{UserID: siswa.ID, Token: token, Role: "siswa"})
		utils.SuccessResponse(c, http.StatusOK, "login berhasil", gin.H{
			"token": token,
			"role":  "siswa",
		})
		return
	}
	utils.ErrorResponse(c, http.StatusUnauthorized, "email tidak ditemukan")
}

func Logout(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "token tidak ditemukan")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "user_id tidak ditemukan")
		return
	}

	role, exists := c.Get("role")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "role tidak ditemukan")
		return
	}

	if err := database.DB.Where("token = ?", tokenString).Delete(&models.Session{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menghapus sesi lama")
		return
	}

	newToken, err := utils.GenerateToken(userID.(uint), role.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal membuat token baru")
		return
	}

	session := models.Session{
		UserID: userID.(uint),
		Token:  newToken,
		Role:   role.(string),
	}

	if err := database.DB.Create(&session).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan sesi baru")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "logout berhasil dan token baru dibuat", gin.H{
		"message":   "token baru udh dibuat",
		"new_token": newToken,
	})
}

func RegisterAdmin(c *gin.Context) {
	var req requests.AdminRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "data tidak valid: "+err.Error())
		return
	}

	if req.Password != req.ConfirmPassword {
		utils.ErrorResponse(c, http.StatusBadRequest, "password dan konfirmasi tidak cocok")
		return
	}

	var existing models.Admin
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "email sudah terdaftar")
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal memproses password")
		return
	}

	admin := models.Admin{
		Nama:     req.Nama,
		Email:    req.Email,
		Password: hashed,
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "gagal menyimpan admin: "+err.Error())
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

	utils.SuccessResponse(c, http.StatusCreated, "registrasi admin berhasil", gin.H{
		"token": token,
		"admin": requests.LoginResponse{ID: admin.ID, Email: admin.Email, Nama: admin.Nama},
	})
}
