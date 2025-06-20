package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterVacationRoutes sets up the routes for vacation requests
func RegisterVacationRoutes(rg *gin.RouterGroup) {
	vacation := rg.Group("/vacations")
	{
		// 1. POST   /vacations           → request a vacation (initial state = "pending")
		vacation.POST("", controllers.CreateVacation)

		// 2. GET    /vacations           → list my vacation requests (optional filters: employee_id, state, from, to)
		vacation.GET("", controllers.ListVacations)

		// 3. GET    /vacations/employees → list subordinates’ vacation requests (supervisors only)
		vacation.GET("/employees", controllers.ListEmployeeVacations)

		// 4. PUT    /vacations/:id/approve → approve a vacation request (state → "approved")
		vacation.PUT("/:id/approve", controllers.ApproveVacation)

		// 5. PUT    /vacations/:id/reject  → reject a vacation request (state → "rejected")
		vacation.PUT("/:id/reject", controllers.RejectVacation)
	}
}
