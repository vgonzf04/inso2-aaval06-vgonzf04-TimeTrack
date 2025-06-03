package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegistrarRutasEmpleado configura todas las rutas REST para empleados
func RegistrarRutasEmpleado(rg *gin.RouterGroup) {
	empleadoGroup := rg.Group("/empleados")
	{
		empleadoGroup.GET("/me", controllers.ObtenerPerfilUsuario)    // Listar datos empleado
		empleadoGroup.GET("/:id", controllers.ObtenerEmpleadoPorID) // Obtener uno
		//empleadoGroup.POST("/", controllers.CrearEmpleado)             // Crear nuevo
		empleadoGroup.PUT("/:id", controllers.ActualizarEmpleado)  // Actualizar
		empleadoGroup.DELETE("/:id", controllers.EliminarEmpleado) // Eliminar
	}
}
