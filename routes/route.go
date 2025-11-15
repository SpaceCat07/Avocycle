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

		// CRUD Kebun
		api.POST("/kebun", controllers.CreateKebun)
		api.GET("/kebun", controllers.GetAllKebun)
		api.GET("/kebun/:id", controllers.GetKebunByID)
		api.PUT("/kebun/:id", controllers.UpdateKebun)
		api.DELETE("/kebun/:id", controllers.DeleteKebun)

		// CRUD Tanaman
		api.POST("/tanaman", middleware.RoleMiddleware("Petani", "Admin"),controllers.CreateTanaman)
		api.GET("/tanaman", controllers.GetAllTanaman)
		api.GET("/tanaman/:id", controllers.GetTanamanByID)
		api.PUT("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.DeleteTanaman)

		petaniRoutes := api.Group("/petani")
		petaniRoutes.Use(middleware.RoleMiddleware("Petani"))
		{
			// CRUD Buah
            petaniRoutes.POST("/buah", controllers.CreateBuah)
            petaniRoutes.GET("/buah", controllers.GetAllBuah)
            petaniRoutes.GET("/buah/id/:id", controllers.GetBuahByID)
			petaniRoutes.GET("/buah/kebun/:id_kebun", controllers.GetBuahByKebun)
            petaniRoutes.PUT("/buah/:id", controllers.UpdateBuah)
            petaniRoutes.DELETE("/buah/:id", controllers.DeleteBuah)

			// CRUD Fase Bunga
			petaniRoutes.POST("/fase-bunga", controllers.CreateFaseBunga)
			petaniRoutes.GET("/fase-bunga", controllers.GetAllFaseBunga)
			petaniRoutes.GET("/fase-bunga/:id", controllers.GetFaseBungaByID)
			petaniRoutes.PUT("/fase-bunga/:id", controllers.UpdateFaseBunga)
			petaniRoutes.DELETE("/fase-bunga/:id", controllers.DeleteFaseBunga)
			petaniRoutes.GET("/fase-bunga/tanaman/:tanaman_id", controllers.GetFaseBungaByTanaman)

			// CRUD Fase Berbuah
			petaniRoutes.POST("/fase-berbuah", controllers.CreateFaseBuah)
			petaniRoutes.GET("/fase-berbuah", controllers.GetAllFaseBuah)
			petaniRoutes.GET("/fase-berbuah/:id", controllers.GetFaseBuahByID)
			petaniRoutes.PUT("/fase-berbuah/:id", controllers.UpdateFaseBuah)
			petaniRoutes.DELETE("/fase-berbuah/:id", controllers.DeleteFaseBuah)
			petaniRoutes.GET("/fase-berbuah/tanaman/:tanaman_id", controllers.GetFaseBuahByTanaman)

			// CRUD Fase Panen
			petaniRoutes.POST("/fase-panen", controllers.CreateFasePanen)
			petaniRoutes.GET("/fase-panen", controllers.GetAllFasePanen)
			petaniRoutes.GET("/fase-panen/:id", controllers.GetFasePanenByID)
			petaniRoutes.PUT("/fase-panen/:id", controllers.UpdateFasePanen)
			petaniRoutes.DELETE("/fase-panen/:id", controllers.DeleteFasePanen)
			petaniRoutes.GET("/fase-panen/tanaman/:tanaman_id", controllers.GetFasePanenByTanaman)
		}
	}

	// cek ketersediaan api
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message" : "Oke"})
	})

	return r
}
