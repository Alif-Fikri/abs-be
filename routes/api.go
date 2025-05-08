package routes

import (
	tc "abs-be/controllers"
	"abs-be/middlewares"

	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	api := r.Group("/api")

	api.POST("/guru/login", tc.LoginGuru)
	api.POST("/admin/login", tc.LoginAdmin)
	api.POST("/wali-kelas/login", tc.LoginWaliKelas)
	api.POST("/logout", middlewares.AuthMiddleware(), tc.Logout)

	api.GET("/dashboard-guru", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("guru"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Guru"})
	})

	api.GET("/dashboard-admin", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Admin"})
	})

	api.GET("/dasboard-walikelas", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Wali Kelas"})

	})

	guru := api.Group("/guru")
	guru.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		guru.GET("/", tc.GetAllGurus)
		guru.GET("/:id", tc.GetGuruByID)
		guru.POST("/", tc.CreateGuru)
		guru.PUT("/:id", tc.UpdateGuru)
		guru.DELETE("/:id", tc.DeleteGuru)

	}

	waliKelas := api.Group("/wali-kelas")
	waliKelas.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"))
	{
		waliKelas.GET("/kelas", tc.GetKelasWaliKelas)
		waliKelas.GET("/siswa", tc.GetSiswaByWaliKelas)
		waliKelas.GET("/daftar", tc.GetAllWaliKelas)
	}

	kelas := api.Group("/kelas")
	kelas.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		kelas.POST("/", tc.CreateKelas)
	}

	mapel := api.Group("/mapel")
	mapel.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		mapel.POST("/", tc.CreateMataPelajaran)
	}

	siswa := api.Group("/siswa")
	siswa.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		siswa.POST("/", tc.CreateSiswa)
		siswa.GET("/", tc.GetSiswaByKelas)
	}

}
