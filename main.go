package main

import (
    "github.com/joho/godotenv"
    "github.com/gin-gonic/gin"

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

    // ── RUTAS PÚBLICAS (sin token JWT) ───────────────────────────────
    routes.RegistrarRutasAuth(router)
    //   └─> Aquí NO ponemos POST /empleados/ (lo reservamos a supervisors)

    // ── RUTAS PROTEGIDAS (requieren JWT válido) ───────────────────────
    protected := router.Group("/")
    protected.Use(middleware.JWTAuth())
    {
        // 5.a) Rutas de empleado (GET, PUT, DELETE). Cualquier usuario autenticado puede llamarlas.
        routes.RegistrarRutasEmpleado(protected)

        // 5.b) Rutas de fichajes (empleado o supervisor)
        routes.RegistrarRutasFichaje(protected)

        // 5.c) Rutas de vacaciones (empleado o supervisor)
        routes.RegistrarRutasVacacion(protected)

		// 5.d) Rutas de dashboard (empleado o supervisor)
		routes.RegistrarRutasDashboard(protected)
    }

    // ── RUTAS EXCLUSIVAS PARA SUPERVISORES ─────────────────────────────
    supervisorOnly := router.Group("/")
    supervisorOnly.Use(middleware.JWTAuth())
    supervisorOnly.Use(middleware.SoloSupervisores())
    {
        // 6.a) Solo un supervisor puede CREAR nuevos empleados
        supervisorOnly.POST("/empleados/", controllers.CrearEmpleado)
    }

    // 7) Arrancar el servidor en el puerto 3000
    router.Run(":3000")
}
