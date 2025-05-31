package routes

import (
	"AppWebPruebaEmpleados/controllers"
	"github.com/gin-gonic/gin"
)

func RegistrarRutasAuth(r *gin.Engine) {
	r.GET("/auth/google", controllers.GoogleLogin)
	r.GET("/auth/google/callback", controllers.GoogleCallback)
}
