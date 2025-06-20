package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterTimecardRoutes sets up the REST routes for timecards (punches)
func RegisterTimecardRoutes(rg *gin.RouterGroup) {
	timecardGroup := rg.Group("/timecards")
	{
		// 1. POST   /timecards           → create a new timecard (clock‐in)
		timecardGroup.POST("", controllers.CreateTimecard)

		// 2. PUT    /timecards/:id/close → close a timecard (clock‐out)
		timecardGroup.PUT("/:id/close", controllers.CloseTimecard)

		// 3. GET    /timecards           → list timecards (with optional query parameters)
		timecardGroup.GET("", controllers.ListTimecards)

		// 4. GET    /timecards/current   → get the currently open timecard
		timecardGroup.GET("/current", controllers.GetCurrentTimecard)
	}
}
