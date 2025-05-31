package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

// RegistrarRutasEmpleado configura todas las rutas REST para empleados
func RegistrarRutasEmpleado(r *gin.Engine) {
	empleadoGroup := r.Group("/empleados")
	{
		empleadoGroup.GET("/", controllers.ObtenerTodosEmpleados)       // Listar todos
		empleadoGroup.GET("/:id", controllers.ObtenerEmpleadoPorID)    // Obtener uno
		empleadoGroup.POST("/", controllers.CrearEmpleado)             // Crear nuevo
		empleadoGroup.PUT("/:id", controllers.ActualizarEmpleado)      // Actualizar
		empleadoGroup.DELETE("/:id", controllers.EliminarEmpleado)     // Eliminar
	}
}
