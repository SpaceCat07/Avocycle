package routes

import (
	"Avocycle/controllers"
	"Avocycle/middleware"

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

		// CRUD Kebun - Tambahan RoleMiddleware untuk Kebun
		api.POST("/kebun", middleware.RoleMiddleware("Petani", "Admin"), controllers.CreateKebun) //1
		api.GET("/kebun", controllers.GetAllKebun)
		api.GET("/kebun/:id", controllers.GetKebunByID)
		api.PUT("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.UpdateKebun) //2
		api.DELETE("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.DeleteKebun) //3

		// CRUD Tanaman
		api.POST("/tanaman", middleware.RoleMiddleware("Petani", "Admin"),controllers.CreateTanaman)
		api.GET("/tanaman", controllers.GetAllTanaman)
		api.GET("/tanaman/:id", controllers.GetTanamanByID)
		api.PUT("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.DeleteTanaman)

	}

	// cek ketersediaan api
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message" : "Oke"})
	})

	return r
}