package middleware

import (
    "fmt"
    "net/http"
    "os"
    "strings"

    "github.com/dgrijalva/jwt-go"
    "github.com/gin-gonic/gin"
)

// JWTAuth valida el token JWT, extrae usuario_id y rol_usuario y los guarda en el contexto.
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1) Leer el header Authorization
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Falta token de autenticación"})
            c.Abort()
            return
        }

        // 2) Debe venir con el formato "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
            c.Abort()
            return
        }
        tokenString := parts[1]

        // 3) Parsear y validar el token usando la clave JWT_SECRET
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Verificar método de firma HMAC
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
            }
            // Retornar la clave para verificar firma
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido o expirado"})
            c.Abort()
            return
        }

        // 4) Extraer las claims
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Claims inválidas"})
            c.Abort()
            return
        }

        // 5) Tomar "sub" como ID de usuario (se almacena como float64)
        sub, ok := claims["sub"].(float64)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Claim 'sub' no válida"})
            c.Abort()
            return
        }
        usuarioID := uint(sub)

        // 6) Tomar "rol" como cadena
        rol, ok := claims["rol"].(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Claim 'rol' no válida"})
            c.Abort()
            return
        }

        // 7) Guardar en el contexto para que los handlers puedan leerlos
        c.Set("usuario_id", usuarioID)
        c.Set("rol_usuario", rol)

        // 8) Continuar hacia el handler
        c.Next()
    }
}
