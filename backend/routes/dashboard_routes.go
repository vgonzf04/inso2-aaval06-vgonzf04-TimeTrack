package routes

import (
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/controllers"
)

// RegisterDashboardRoutes registers all /dashboard endpoints under a RouterGroup.
func RegisterDashboardRoutes(rg *gin.RouterGroup) {
	dashboard := rg.Group("/dashboard")
	{
		// 1. Get worked hours per employee over a period (optional employee_id filter)
		dashboard.GET("/hours-period", controllers.HoursWorkedByPeriod)

		// 2. Count open and closed punches for a given day (supervisors only)
		dashboard.GET("/checkins-day", controllers.PunchesByDay)

		// 3. Group vacation requests by state (supervisors only)
		dashboard.GET("/vacations-by-status", controllers.VacationsByState)
	}
}
