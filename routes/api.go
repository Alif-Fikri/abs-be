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
}
