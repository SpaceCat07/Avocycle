package routes

import (
	"Avocycle/controllers"
	"Avocycle/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	// ======================
	// 📦 Public Routes
	// ======================
	api := r.Group("/api/v1")
	{
		api.POST("register/petani", controllers.RegisterPetani)
		api.POST("register/pembeli", controllers.RegisterPembeli)
		api.POST("register/admin", controllers.RegisterAdmin)
		api.POST("login", controllers.Login)
	}

	// ======================
	// 🔐 Protected Routes
	// ======================
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())

	// hanya admin
	protected.GET("/admin/dashboard", middleware.RoleMiddleware("admin"), controllers.AdminDashboard)

	// hanya petani
	protected.GET("/petani/dashboard", middleware.RoleMiddleware("petani"), controllers.PetaniDashboard)

	// hanya pembeli
	protected.GET("/pembeli/dashboard", middleware.RoleMiddleware("pembeli"), controllers.PembeliDashboard)

	// ======================
	// 🚀 Test Endpoint
	// ======================
	r.GET("/oke", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Oke"})
	})

	return r
}
