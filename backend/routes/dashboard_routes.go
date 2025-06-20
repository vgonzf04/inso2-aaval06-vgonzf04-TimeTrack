package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

// RegisterDashboardRoutes registers all /dashboard endpoints under a RouterGroup.
func RegisterDashboardRoutes(rg *gin.RouterGroup) {
	dashboard := rg.Group("/dashboard")
	{
		// 1. Get worked hours per employee over a period (optional empleado_id filter)
		dashboard.GET("/hours-period", controllers.HoursWorkedInPeriod)

		// 2. Count open and closed check-ins for a given day
		dashboard.GET("/checkins-day", controllers.CheckinsByDay)

		// 3. Group vacation requests by status
		dashboard.GET("/vacations-by-status", controllers.VacationsByStatus)
	}
}
