package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTAuth validates the JWT token, extracts user_id and user_role, and stores them in the context.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Read the "token" cookie
		authCookie, err := c.Request.Cookie("token")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(c.Writer, "cookie not found", http.StatusBadRequest)
			default:
				log.Println(err)
				http.Error(c.Writer, "server error", http.StatusInternalServerError)
			}
			return
		}

		tokenString := authCookie.Value

		// 2) Parse and validate the token using JWT_SECRET
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify HMAC signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the key for signature verification
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 3) Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		// 4) Get "sub" as user ID (stored as float64)
		sub, ok := claims["sub"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 'sub' claim"})
			c.Abort()
			return
		}
		userID := uint(sub)

		// 5) Get "rol" as user role
		role, ok := claims["rol"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 'role' claim"})
			c.Abort()
			return
		}

		// 6) Store in context for handlers
		c.Set("usuario_id", userID)
		c.Set("rol_usuario", role)

		// 7) Continue to handler
		c.Next()
	}
}