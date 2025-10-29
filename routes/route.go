package routes

import (
	"Avocycle/controllers"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	// url api v1
	api := r.Group("/api/v1")
	{
		api.POST("register/petani", controllers.ManualRegisterPetani)
		api.POST("register/pembeli", controllers.ManualRegisterPembeli)
		api.POST("login", controllers.ManualLogin)

		api.GET("auth/:provider/petani", controllers.RedirectHandlerPetani)
		api.GET("auth/:provider/callback/petani", controllers.CallbackHandlerPetani)

		api.GET("auth/:provider/pembeli", controllers.RedirectHandlerPembeli)
		api.GET("auth/:provider/callback/pembeli", controllers.CallbackHandlerPembeli)

		// CRUD Kebun
		api.POST("/kebun", controllers.CreateKebun)
		api.GET("/kebun", controllers.GetAllKebun)
		api.GET("/kebun/:id", controllers.GetKebunByID)
		api.PUT("/kebun/:id", controllers.UpdateKebun)
		api.DELETE("/kebun/:id", controllers.DeleteKebun)

		// CRUD Tanaman
		api.POST("/tanaman", controllers.CreateTanaman)
		api.GET("/tanaman", controllers.GetAllTanaman)
		api.GET("/tanaman/:id", controllers.GetTanamanByID)
		api.PUT("/tanaman/:id", controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", controllers.DeleteTanaman)

	}

	// cek ketersediaan api
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message" : "Oke"})
	})

	return r
}