package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

func RegistrarRutasDashboard(rg *gin.RouterGroup) {
	dash := rg.Group("/dashboard")
	{
		// 1. Horas trabajadas por empleado en un período (filtro opcional por empleado o supervisor)
		dash.GET("/horas-periodo", controllers.HorasTrabajadasPorPeriodo)

		// 2. Número de fichajes abiertos y cerrados en un día dado
		dash.GET("/fichajes-dia", controllers.FichajesPorDia)

		// 3. Solicitudes de vacaciones por estado
		dash.GET("/vacaciones-por-estado", controllers.VacacionesPorEstado)

	}
}
