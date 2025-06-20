package routes

import (
	"AppWebPruebaEmpleados/controllers"
	"AppWebPruebaEmpleados/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers all the authentication-related endpoints.
func RegisterAuthRoutes(r *gin.Engine) {
	// Start OAuth2 flow with Google
	r.GET("/auth/google", controllers.GoogleLogin)
	// OAuth2 callback endpoint
	r.GET("/auth/google/callback", controllers.GoogleCallback)
	// Returns the authenticated user's full profile
	r.GET("/auth/me", middleware.JWTAuth(), controllers.GetAuthenticatedUser)
	// Returns only the user's role (for quick checks)
	r.GET("/me", controllers.GetUserRole)
	// Logs the user out by clearing the JWT cookie
	r.POST("/auth/logout", controllers.Logout)
}
