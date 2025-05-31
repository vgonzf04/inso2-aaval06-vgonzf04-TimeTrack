package main

import (
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/middleware"
	"AppWebPruebaEmpleados/controllers"
	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/routes"
)

func main() {
	// Cagar .env
	godotenv.Load()
	// Conectar a la base de datos
	config.ConectarBD()

	// Autenticación con google
	config.InitGoogleOAuth()

	// Crear router con Gin
	router := gin.Default()

	// Registrar rutas autenticacion
	routes.RegistrarRutasAuth(router)

	// Registrar las rutas de empleados
	routes.RegistrarRutasEmpleado(router)

	// Registrar las rutas de fichajes
    routes.RegistrarRutasFichaje(router)

	supervisorOnly := router.Group("/")
	supervisorOnly.Use(middleware.auth())
	supervisorOnly.Use(middleware.soloSupervisor())
	{
		// 4.a) ÚNICAMENTE un supervisor puede CREAR nuevos empleados:
		supervisorOnly.POST("/empleados/", controllers.CrearEmpleado)
	}
	Registrar las rutas de vacaciones
	routes.RegistrarRutasVacacion(router)

	// Iniciar el servidor web en el puerto 8080
	router.Run(":3000")
}
