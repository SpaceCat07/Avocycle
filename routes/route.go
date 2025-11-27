package routes

import (
	"Avocycle/controllers"
	"Avocycle/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	// setting cors
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
	}))

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
		api.GET("/tanaman/by-kebun/:id_kebun", controllers.GetTanamanByKebunID)
		api.PUT("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.DeleteTanaman)

		// log penyakit tanaman
		api.GET("/Log-Penyakit-Tanaman", controllers.GetAllLogPenyakit)
		api.GET("/Log-Penyakit-Tanaman/:id", controllers.GetLogPenyakitById)
		api.GET("/Log-Penyakit-Tanaman/Tanaman/:id_tanaman", controllers.GetLogPenyakitByTanamanId)

		petaniRoutes := api.Group("/petani")
		petaniRoutes.Use(middleware.RoleMiddleware("Petani"))
		{
			// CRUD Buah
            petaniRoutes.POST("/buah", controllers.CreateBuah)
            petaniRoutes.GET("/buah", controllers.GetAllBuah)
            petaniRoutes.GET("/buah/:id", controllers.GetBuahByID)
			petaniRoutes.GET("/buah/by-tanaman/:id_kebun", controllers.GetBuahByKebun)
            petaniRoutes.PUT("/buah/:id", controllers.UpdateBuah)
            petaniRoutes.DELETE("/buah/:id", controllers.DeleteBuah)
		}

		petaniAdminRoutes := api.Group("/petamin")
		petaniAdminRoutes.Use(middleware.RoleMiddleware("Petani", "Admin"))
		{
			// deteksi penyakit tanaman
			petaniAdminRoutes.POST("/penyakit", controllers.ClassifyPenyakit)
		}
	}

	// cek ketersediaan api
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message" : "Oke"})
	})

	return r
}