package main

import (
	"abs-be/database"
	"abs-be/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.Konek()

	r := gin.Default()
	r.Use(cors.Default())
	routes.Api(r)

	if err := r.Run(":8080"); err != nil {
		panic("gagal menjalankan server: " + err.Error())
	}
}
