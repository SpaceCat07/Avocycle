package main

import (
	"Avocycle/config"
	"Avocycle/routes"
	// "fmt"

	// "github.com/gin-gonic/gin"
)

func main() {
	// ini untuk seeder nanti


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