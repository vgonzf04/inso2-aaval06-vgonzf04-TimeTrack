package main

import (
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
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

	// Registrar las rutas de vacaciones
	routes.RegistrarRutasVacacion(router)

	// Iniciar el servidor web en el puerto 8080
	router.Run(":3000")
}
