package routes

import (
	"AppWebPruebaEmpleados/controllers"
	"AppWebPruebaEmpleados/middleware"

	"github.com/gin-gonic/gin"
)

func RegistrarRutasAuth(r *gin.Engine) {
	r.GET("/auth/google", controllers.GoogleLogin)
	r.GET("/auth/google/callback", controllers.GoogleCallback)
	r.GET("/auth/me", middleware.JWTAuth(), controllers.ObtenerUsuarioAutenticado)
	r.GET("/me", controllers.Me)
	r.POST("/auth/logout", controllers.Logout)
}
