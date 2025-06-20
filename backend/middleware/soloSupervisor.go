package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// OnlySupervisors aborts the request if the user's role is not "supervisor".
func OnlySupervisors() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1) Retrieve "rol_usuario" from context
        roleIface, exists := c.Get("rol_usuario")
        if !exists {
            // If the role is missing, respond with an internal error
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve user role from context"})
            c.Abort()
            return
        }

        // 2) Assert the value is a string and equals "supervisor"
        role, ok := roleIface.(string)
        if !ok || role != "supervisor" {
            // If not a supervisor, deny access
            c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: supervisors only"})
            c.Abort()
            return
        }

        // 3) Proceed to the next handler
        c.Next()
    }
}
