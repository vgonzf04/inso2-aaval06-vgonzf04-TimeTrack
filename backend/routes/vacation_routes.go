package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegistrarRutasVacacion configura las rutas para las solicitudes de vacaciones
func RegistrarRutasVacacion(rg *gin.RouterGroup) {
	vaca := rg.Group("/vacaciones")
	{
		// 1. Solicitar vacaciones (estado inicial = "pendiente")
		vaca.POST("", controllers.CrearVacacion)

		// 2. Obtener lista de solicitudes (posibles filtros: empleado_id, estado, desde, hasta)
		vaca.GET("", controllers.ListarVacaciones)
		vaca.GET("/empleados", controllers.ListarVacacionesEmpleados)

		// 3. Aprobar una solicitud: cambia estado a "aprobada"
		vaca.PUT("/:id/aprobar", controllers.AprobarVacacion)

		// 4. Rechazar una solicitud: cambia estado a "rechazada"
		vaca.PUT("/:id/rechazar", controllers.RechazarVacacion)
	}
}
