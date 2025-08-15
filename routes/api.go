package routes

import (
	tc "abs-be/controllers"
	"abs-be/middlewares"

	"github.com/gin-gonic/gin"
)

func Api(r *gin.Engine) {
	api := r.Group("/api")
	api.POST("/login", tc.LoginAutoRole)    // dipake
	api.POST("/guru/login", tc.LoginGuru)   // ga dipake (cuma testing)
	api.POST("/admin/login", tc.LoginAdmin) // ga dipake (cuma testing)
	api.POST("/admin/register", tc.RegisterAdmin)
	api.POST("/wali-kelas/login", tc.LoginWaliKelas) // ga dipake (cuma testing)
	api.POST("/logout", middlewares.AuthMiddleware(), tc.Logout)

	api.GET("/dashboard-guru", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("guru"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Guru"}) // ga dipake (cuma testing)
	})

	api.GET("/dashboard-admin", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Admin"}) // ga dipake (cuma testing)
	})

	api.GET("/dashboard-walikelas", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Wali Kelas"}) // ga dipake (cuma testing)

	})

	api.GET("/dashboard-siswa", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("siswa"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "dashboard Siswa"}) // ga dipake (cuma testing)
	})

	guru := api.Group("/guru")
	guru.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	{
		guru.GET("/", tc.GetAllGurus)
		guru.GET("/:id", tc.GetGuruByID)
		guru.POST("/", tc.CreateGuru)
		guru.PUT("/:id", tc.UpdateGuru)
		guru.DELETE("/:id", tc.DeleteGuru)
		guru.POST("/assign-mapel", tc.AssignMapelKelas)
		guru.GET("/assign-mapel/", tc.ListAssignMapelKelas)
		guru.PUT("/assign-mapel/:id", tc.UpdateAssignMapelKelas)
		guru.DELETE("/assign-mapel/:id", tc.DeleteAssignMapelKelas)
		guru.POST("/assign-walikelas", tc.AssignWaliKelas)
		guru.POST("/unassign-walikelas", tc.UnassignWaliKelas)
		guru.GET("/wali-kelas/daftar", tc.GetAllWaliKelas)
	}

	pengajar := api.Group("/guru")
	pengajar.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("guru"))
	{
		pengajar.GET("/list-kelas", tc.GetPengajaranGuru)
	}

	waliKelas := api.Group("/wali-kelas")
	waliKelas.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("wali_kelas"))
	{
		waliKelas.GET("/siswa", tc.GetSiswaByWaliKelas)
		waliKelas.GET("/daftar", tc.GetAllWaliKelas)
		waliKelas.GET("/list-kelas", tc.GetKelasWali)
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
		siswa.POST("/assign-kelas", tc.AssignSiswaToKelas)
		siswa.POST("/unassign-kelas", tc.UnassignSiswaFromKelas)
		siswa.POST("/assign-mapel", tc.AssignSiswaToMapel)
		siswa.POST("/unassign-mapel", tc.UnassignSiswaFromMapel)
	}

	siswaaja := api.Group("/siswa")
	siswaaja.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("siswa"))
	{
		siswaaja.GET("/profil", tc.GetProfilSiswa)
		siswaaja.GET("/absensi", tc.GetAbsensiSiswa)
	}

	absensi := api.Group("/absensi")
	absensi.Use(middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin", "guru", "wali_kelas"))
	{
		absensi.POST("/", tc.CreateAbsensiSiswa)
		absensi.GET("/", tc.GetAbsensi)
		absensi.GET("/:id", tc.GetAbsensiByID)
		absensi.PUT("/:id", tc.UpdateAbsensiSiswa)
		absensi.DELETE("/:id", tc.DeleteAbsensiSiswa)
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
