package main

import (
    _ "Avocycle/docs"
	"Avocycle/config"
	"Avocycle/routes"

	"github.com/joho/godotenv"
	// "fmt"
	// "github.com/gin-gonic/gin"
)

// @title       Avocycle API
// @version     1.0
// @description Dokumentasi API Avocycle (Kebun CRUD, Pagination, dan Response Format)
// @host        localhost:2005
// @BasePath 	/api/v1
// @schemes     http
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer <your-token>"
// @security Bearer
func main() {
	// ini untuk seeder nanti

	// initialize gocial
	config.InitGocial()

	godotenv.Load()
	// connect to postgres
	postsql, err := config.DbConnect()
	if err != nil {
		panic("Failed to get database connection: " + err.Error())
	}

	// Get underlying sql.DB to close the connection
    sqlDB, err := postsql.DB()
    if err != nil {
        panic("Failed to get underlying database connection: " + err.Error())
    }

	defer sqlDB.Close()

	router := routes.InitRoutes()

	router.Run(":2005")
}