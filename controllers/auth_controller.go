package controllers

import (
	"context"
	"net/http"

	"AppWebPruebaEmpleados/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	gservice "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

func GoogleLogin(c *gin.Context) {
	url := config.GoogleOAuthConfig.AuthCodeURL("random-state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	code := c.Query("code")

	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al intercambiar token"})
		return
	}

	svc, err := gservice.NewService(
		context.Background(),
		option.WithTokenSource(config.GoogleOAuthConfig.TokenSource(context.Background(), token)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear servicio de Google"})
		return
	}

	userInfo, err := gservice.NewUserinfoV2MeService(svc).Get().Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener información del usuario"})
		return
	}

	// Registrar el usuario si no existe, o generar un token
	c.JSON(http.StatusOK, gin.H{
		"nombre": userInfo.Name,
		"email":  userInfo.Email,
		"foto":   userInfo.Picture,
	})
}
