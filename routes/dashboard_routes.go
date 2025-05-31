package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

func RegistrarRutasDashboard(r *gin.Engine) {
	dash := r.Group("/dashboard")
	{
		// 1. Horas trabajadas por empleado en un período (filtro opcional por empleado o supervisor)
		dash.GET("/horas-periodo", controllers.HorasTrabajadasPorPeriodo)

		// 2. Promedio de horas trabajadas en un período, por empleado
		dash.GET("/promedio-horas", controllers.PromedioHorasPorPeriodo)

		// 3. Número de fichajes abiertos y cerrados en un día dado
		dash.GET("/fichajes-dia", controllers.FichajesPorDia)

		// 4. Solicitudes de vacaciones por estado
		dash.GET("/vacaciones-por-estado", controllers.VacacionesPorEstado)

		// 5. Días de vacaciones consumidos por empleado en un rango
		dash.GET("/vacaciones-consumidas-rango", controllers.VacacionesConsumidasPorRango)
	}
}
