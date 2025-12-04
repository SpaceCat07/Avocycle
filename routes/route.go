package routes

import (
	"Avocycle/controllers"
	_ "Avocycle/docs"
	"Avocycle/middleware"
	"time"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	// setting cors
	r.Use(cors.New(cors.Config{
		AllowOrigins:	  []string{
							    "https://avocycle.shop",
							    "https://www.avocycle.shop",
							    "http://localhost:3000", // development
							},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// url api v1
	api := r.Group("/api/v1")
	{
		api.POST("register/petani", controllers.ManualRegisterPetani)
		api.POST("register/pembeli", controllers.ManualRegisterPembeli)
		api.POST("login", controllers.ManualLogin)

		// Pembeli
		api.GET("auth/google/pembeli", controllers.RedirectHandlerPembeli)
		api.GET("auth/:provider/callback/pembeli", controllers.CallbackHandlerPembeli)
		api.POST("auth/google/complete/pembeli", controllers.CompleteGooglePembeli)

		// Petani
		api.GET("auth/google/petani", controllers.RedirectHandlerPetani)
		api.GET("auth/:provider/callback/petani", controllers.CallbackHandlerPetani)
		api.POST("auth/google/complete/petani", controllers.CompleteGooglePetani)

		// CRUD Kebun - Tambahan RoleMiddleware untuk Kebun
		api.POST("/kebun", middleware.RoleMiddleware("Petani", "Admin"), controllers.CreateKebun) //1
		api.GET("/kebun", controllers.GetAllKebun)
		api.GET("/kebun/:id", controllers.GetKebunByID)
		api.PUT("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.UpdateKebun)    //2
		api.DELETE("/kebun/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.DeleteKebun) //3

		// CRUD Tanaman
		api.POST("/tanaman", middleware.RoleMiddleware("Petani", "Admin"), controllers.CreateTanaman)
		api.GET("/tanaman", controllers.GetAllTanaman)
		api.GET("/tanaman/:id", controllers.GetTanamanByID)
		api.GET("/tanaman/by-kebun/:id_kebun", controllers.GetTanamanByKebunID)
		api.PUT("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.UpdateTanaman)
		api.DELETE("/tanaman/:id", middleware.RoleMiddleware("Petani", "Admin"), controllers.DeleteTanaman)

		// log penyakit tanaman
		api.GET("/Log-Penyakit-Tanaman", controllers.GetAllLogPenyakit)
		api.GET("/Log-Penyakit-Tanaman/:id", controllers.GetLogPenyakitById)
		api.GET("/Log-Penyakit-Tanaman/Tanaman/:id_tanaman", controllers.GetLogPenyakitByTanamanId)

		// get fase bunga
		api.GET("/fase-bunga", controllers.GetAllFaseBunga)
		api.GET("/fase-bunga/:id", controllers.GetFaseBungaByID)
		api.GET("/fase-bunga/tanaman/:tanaman_id", controllers.GetFaseBungaByTanaman)

		// get fase berbuah
		api.GET("/fase-berbuah", controllers.GetAllFaseBuah)
		api.GET("/fase-berbuah/:id", controllers.GetFaseBuahByID)
		api.GET("/fase-berbuah/tanaman/:tanaman_id", controllers.GetFaseBuahByTanaman)

		// get fase panen
		api.GET("/fase-panen", controllers.GetAllFasePanen)
		api.GET("/fase-panen/:id", controllers.GetFasePanenByID)
		api.GET("/fase-panen/tanaman/:tanaman_id", controllers.GetFasePanenByTanaman)
		

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

			// statistics
			petaniRoutes.GET("/count-all-tanaman", controllers.CountAllPohon)
			petaniRoutes.GET("/count-tanaman-sakit", controllers.CountTanamanDiseased)
			petaniRoutes.GET("/count-tanaman-siap-panen", controllers.CountSiapPanen)
			petaniRoutes.GET("/count-tanaman-tiap-minggu", controllers.GetWeeklyPanenLast6Weeks)

			// CRUD Fase Bunga
			petaniRoutes.POST("/fase-bunga", controllers.CreateFaseBunga)
			petaniRoutes.PUT("/fase-bunga/:id", controllers.UpdateFaseBunga)
			petaniRoutes.DELETE("/fase-bunga/:id", controllers.DeleteFaseBunga)

			// CRUD Fase Berbuah
			petaniRoutes.POST("/fase-berbuah", controllers.CreateFaseBuah)
			petaniRoutes.PUT("/fase-berbuah/:id", controllers.UpdateFaseBuah)
			petaniRoutes.DELETE("/fase-berbuah/:id", controllers.DeleteFaseBuah)

			// CRUD Fase Panen
			petaniRoutes.POST("/fase-panen", controllers.CreateFasePanen)
			petaniRoutes.PUT("/fase-panen/:id", controllers.UpdateFasePanen)
			petaniRoutes.DELETE("/fase-panen/:id", controllers.DeleteFasePanen)
		}

		petaniAdminRoutes := api.Group("/petamin")
		petaniAdminRoutes.Use(middleware.RoleMiddleware("Petani", "Admin"))
		{
			// deteksi penyakit tanaman
			petaniAdminRoutes.POST("/penyakit/:id_tanaman", controllers.ClassifyPenyakit)
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
		ctx.JSON(200, gin.H{"message": "Oke"})
	})

	return r
}
