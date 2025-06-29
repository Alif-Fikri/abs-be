package routes

import (
	tc "abs-be/controllers"
	"abs-be/middlewares"

	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/login", tc.LoginAll)
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

	api.GET("/dashboard-walikelas", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Wali Kelas"})

	})

	api.GET("/dashboard-siswa", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("siswa"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Siswa"})
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
		kelas.GET("/", tc.GetAllKelas)
		kelas.GET("/:id", tc.GetKelasByID)
		kelas.PUT("/:id", tc.UpdateKelas)
		kelas.DELETE("/:id", tc.DeleteKelas)
	}

	mapel := api.Group("/mapel")
	mapel.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		mapel.POST("/", tc.CreateMataPelajaran)
		mapel.GET("/", tc.GetAllMataPelajaran)
		mapel.GET("/:id", tc.GetMataPelajaranByID)
		mapel.PUT("/:id", tc.UpdateMataPelajaran)
		mapel.DELETE("/:id", tc.DeleteMataPelajaran)
	}

	siswa := api.Group("/siswa")
	siswa.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		siswa.POST("/", tc.CreateSiswa)
		siswa.GET("/", tc.GetAllSiswa)
		siswa.GET("/kelas/:id", tc.GetSiswaByKelas)
		siswa.GET("/:id", tc.GetSiswaByID)
		siswa.PUT("/:id", tc.UpdateSiswa)
		siswa.DELETE("/:id", tc.DeleteSiswa)

	}

	siswaaja := api.Group("/siswa")
	siswaaja.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("siswa"))
	{
		siswaaja.GET("/profil", tc.GetProfilSiswa)
		siswaaja.GET("/absensi", tc.GetAbsensiSiswa)
	}

	absensi := api.Group("/absensi")
	absensi.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("guru", "wali_kelas"))
	{
		absensi.POST("/", tc.CreateAbsensiSiswa)
		absensi.GET("/list/mapel", tc.ListStudentsForMapel)
		absensi.GET("/list/kelas", tc.ListStudentsForKelas)
		absensi.GET("/rekap/mapel", tc.RecapAbsensiMapel)
		absensi.GET("/rekap/kelas", tc.RecapAbsensiKelas)
	}

	todo := api.Group("/todo")
	todo.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "guru", "wali_kelas"))
	{
		todo.POST("/", tc.CreateTodo)
		todo.GET("/", tc.GetTodosByTanggal) // query?tanggal=YYYY-MM-DD
		todo.PUT("/:id/status", tc.UpdateTodoStatus)
		todo.DELETE("/:id", tc.DeleteTodo)
	}
}
