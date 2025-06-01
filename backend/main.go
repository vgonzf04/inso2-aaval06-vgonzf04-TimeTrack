package main

import (
	"time"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/controllers"
	"AppWebPruebaEmpleados/middleware"
	"AppWebPruebaEmpleados/routes"
)

func main() {
	// 1) Cargar variables de entorno (.env)
	godotenv.Load()

	// 2) Conectar a la BD y hacer AutoMigrate
	config.ConectarBD()

	// 3) Inicializar Google OAuth
	config.InitGoogleOAuth()

	// 4) Crear router Gin
	router := gin.Default()

	// ✅ 5) Activar CORS para permitir peticiones desde el frontend (puerto 3001)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── RUTAS PÚBLICAS (sin token JWT) ───────────────────────────────
	routes.RegistrarRutasAuth(router)

	// ── RUTAS PROTEGIDAS (requieren JWT válido) ───────────────────────
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		routes.RegistrarRutasEmpleado(protected)
		routes.RegistrarRutasFichaje(protected)
		routes.RegistrarRutasVacacion(protected)
		routes.RegistrarRutasDashboard(protected)
	}

	// ── RUTAS EXCLUSIVAS PARA SUPERVISORES ─────────────────────────────
	supervisorOnly := router.Group("/")
	supervisorOnly.Use(middleware.JWTAuth())
	supervisorOnly.Use(middleware.SoloSupervisores())
	{
		supervisorOnly.POST("/empleados/", controllers.CrearEmpleado)
	}

	// 7) Arrancar el servidor en el puerto 3000
	router.Run(":3000")

	router.NoRoute(func(c *gin.Context) {
		fmt.Println("🚨 Ruta no encontrada:", c.Request.Method, c.Request.URL.Path)
	})
}
