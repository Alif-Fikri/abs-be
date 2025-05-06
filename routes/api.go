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

	guru := api.Group("/guru")
	guru.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		guru.GET("/", tc.GetAllGurus)
		guru.GET("/:id", tc.GetGuruByID)
		guru.POST("/", tc.CreateGuru)
		guru.PUT("/:id", tc.UpdateGuru)
		guru.DELETE("/:id", tc.DeleteGuru)
	}

	api.GET("/dashboard-guru", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("guru"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Guru"})
	})

	waliKelas := api.Group("/wali-kelas")
	waliKelas.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"))
	{
		waliKelas.GET("/kelas", tc.GetKelasWaliKelas)   
		waliKelas.GET("/siswa", tc.GetSiswaByWaliKelas) 
		waliKelas.GET("/daftar", tc.GetAllWaliKelas)
	}

}
