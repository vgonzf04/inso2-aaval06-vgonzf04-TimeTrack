package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterEmployeeRoutes sets up all REST routes for employees
func RegisterEmployeeRoutes(rg *gin.RouterGroup) {
	employeeGroup := rg.Group("/employees")
	{
		// GET /employees/me       → return authenticated user’s profile
		employeeGroup.GET("/me", controllers.GetUserProfile)

		// GET /employees/:id      → return a specific employee by ID
		employeeGroup.GET("/:id", controllers.GetEmployeeByID)

		// PUT /employees/:id      → update an existing employee
		employeeGroup.PUT("/:id", controllers.UpdateEmployee)

		// DELETE /employees/:id   → delete an employee
		employeeGroup.DELETE("/:id", controllers.DeleteEmployee)
	}
}
