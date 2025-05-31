package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

// RegistrarRutasFichaje configura las rutas REST para fichajes
func RegistrarRutasFichaje(r *gin.Engine) {
	fichajeGroup := r.Group("/fichajes")
	{
		// 1. Un endpoint para crear un fichaje (entrada)
		fichajeGroup.POST("/", controllers.CrearFichaje)

		// 2. Un endpoint para marcar salida (actualizar horaSalida)
		fichajeGroup.PUT("/:id/cerrar", controllers.CerrarFichaje)

		// 3. Un endpoint para listar fichajes (con posibles query params)
		fichajeGroup.GET("/", controllers.ListarFichajes)
	}
}
