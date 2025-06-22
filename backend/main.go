package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/controllers"
	"AppWebPruebaEmpleados/middleware"
	"AppWebPruebaEmpleados/routes"
)

func main() {
	// 1) Cargar variables de entorno
	godotenv.Load()

	// 2) Conectar a la BD y AutoMigrate
	config.ConnectDB()

	// 3) Inicializar Google OAuth
	config.InitGoogleOAuth()

	// 4) Crear router Gin
	router := gin.Default()

	// 5) CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── RUTAS PÚBLICAS ──────────────────────────────────
	routes.RegisterAuthRoutes(router)

	// ── RUTAS PROTEGIDAS (requieren JWT) ───────────────
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		routes.RegisterEmployeeRoutes(protected)
		routes.RegisterTimecardRoutes(protected)   // <--- Aquí añadimos las rutas de timecards
		routes.RegisterVacationRoutes(protected)
		routes.RegisterDashboardRoutes(protected)
	}

	// ── RUTAS SÓLO SUPERVISORES ─────────────────────────
	supervisorOnly := router.Group("/")
	supervisorOnly.Use(middleware.JWTAuth(), middleware.OnlySupervisors())
	{
		// POST /employees → crear nuevo empleado
		supervisorOnly.POST("/employees", controllers.CreateEmployee)
	}

	// Manejo de ruta no encontrada
	router.NoRoute(func(c *gin.Context) {
		fmt.Println("🚨 Ruta no encontrada:", c.Request.Method, c.Request.URL.Path)
		c.JSON(404, gin.H{"error": "Not Found"})
	})

	// 7) Arrancar servidor
	router.Run(":3000")
}
