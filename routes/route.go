package routes

import (
	_ "Avocycle/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"Avocycle/controllers"
	"Avocycle/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		api.POST("/kebun", middleware.RoleMiddleware("Petani", "Admin"), controllers.CreateKebun)
		api.GET("/kebun", controllers.GetAllKebun)
		api.GET("/kebun/:id", controllers.GetKebunByID)
		api.PUT("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.UpdateKebun)
		api.DELETE("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.DeleteKebun)

		// CRUD Tanaman
		api.POST("/tanaman", middleware.RoleMiddleware("Petani", "Admin"),controllers.CreateTanaman)
		api.GET("/tanaman", controllers.GetAllTanaman)
		api.GET("/tanaman/:id", controllers.GetTanamanByID)
		api.GET("/tanaman/kebun/:id_kebun", controllers.GetTanamanByKebunID)
		api.PUT("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"),controllers.DeleteTanaman)

		// log penyakit tanaman
		api.GET("/Log-Penyakit-Tanaman", controllers.GetAllLogPenyakit)
		api.GET("/Log-Penyakit-Tanaman/:id", controllers.GetLogPenyakitById)
		api.GET("/Log-Penyakit-Tanaman/Tanaman/:id_tanaman", controllers.GetLogPenyakitByTanamanId)

		// middleware khusus petani
		petaniRoutes := api.Group("/petani")
		petaniRoutes.Use(middleware.RoleMiddleware("Petani"))
		{
			// CRUD Buah
            petaniRoutes.POST("/buah", controllers.CreateBuah)
            petaniRoutes.GET("/buah", controllers.GetAllBuah)
            petaniRoutes.GET("/buah/:id", controllers.GetBuahByID)
			petaniRoutes.GET("/buah/kebun/:id_kebun", controllers.GetBuahByKebun)
            petaniRoutes.PUT("/buah/:id", controllers.UpdateBuah)
            petaniRoutes.DELETE("/buah/:id", controllers.DeleteBuah)
		}

		petaniAdminRoutes := api.Group("/petamin")
		petaniAdminRoutes.Use(middleware.RoleMiddleware("Petani", "Admin"))
		{
			// deteksi penyakit tanaman
			petaniAdminRoutes.POST("/penyakit", controllers.ClassifyPenyakit)
		}

		// middleware khusus Pembeli
		pembeliRoutes := api.Group("/pembeli")
		pembeliRoutes.Use(middleware.RoleMiddleware("Pembeli"))
		{
			// CRUD Booking
			pembeliRoutes.POST("/booking", controllers.CreateBooking)
			pembeliRoutes.GET("/booking", controllers.GetAllBooking)
			pembeliRoutes.GET("/booking/:id", controllers.GetBookingByID)
			pembeliRoutes.PUT("/booking/:id", controllers.UpdateBooking)
			pembeliRoutes.DELETE("/booking/:id", controllers.DeleteBooking)
			pembeliRoutes.GET("/booking/user/:user_id", controllers.GetBookingByUserID)
		}

	}

	// cek ketersediaan api
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message" : "Oke"})
	})

	return r
}