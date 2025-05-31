package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// SoloSupervisores aborta la petición si rol_usuario != "supervisor"
func SoloSupervisores() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1) Sacar de contexto el valor “rol_usuario”
        rolIface, exists := c.Get("rol_usuario")
        if !exists {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo leer rol del usuario"})
            c.Abort()
            return
        }

        // 2) Convertir a string
        rol, ok := rolIface.(string)
        if !ok || rol != "supervisor" {
            c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores"})
            c.Abort()
            return
        }

        // 3) Si es supervisor, continuar
        c.Next()
    }
}
