package routes

import (
	"AppWebPruebaEmpleados/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterTimecardRoutes sets up the REST routes for timecards (punches)
func RegisterTimecardRoutes(rg *gin.RouterGroup) {
	timecardGroup := rg.Group("/timecards")
	{
		// 1. POST   /timecards           → create a new time entry (clock-in)
		timecardGroup.POST("", controllers.CreateTimeEntry)

		// 2. PUT    /timecards/:id/close → close a time entry (clock-out)
		timecardGroup.PUT("/:id/close", controllers.CloseTimeEntry)

		// 3. GET    /timecards           → list time entries (with optional query parameters)
		timecardGroup.GET("", controllers.ListTimeEntries)

		// 4. GET    /timecards/current   → get the currently open time entry
		timecardGroup.GET("/current", controllers.GetCurrentTimeEntry)
	}
}
