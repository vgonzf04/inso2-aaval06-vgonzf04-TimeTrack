package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	gservice "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

func Me(c *gin.Context) {
    // 1) Leemos la cookie “token”
    tokenString, err := c.Cookie("token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
        return
    }

    // 2) Parseamos y validamos el JWT con la misma secret
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    if err != nil || !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
        return
    }

    // 3) Extraemos los claims (mapa) y sacamos “rol”
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno leyendo token"})
        return
    }
    rolInterfaz, ok := claims["rol"]
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No se encontró rol en claims"})
        return
    }
    rol, ok := rolInterfaz.(string)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de rol inválido en token"})
        return
    }

    // 4) Devolvemos JSON con el rol
    c.JSON(http.StatusOK, gin.H{
        "rol": rol,
    })
}


func GoogleLogin(c *gin.Context) {
	url := config.GoogleOAuthConfig.AuthCodeURL("random-state", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	code := c.Query("code")

	// 1) Intercambiar el código por un token de Google
	tokenGoogle, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al intercambiar token con Google"})
		return
	}

	// 2) Crear servicio de Google para obtener información del usuario
	svc, err := gservice.NewService(
		context.Background(),
		option.WithTokenSource(config.GoogleOAuthConfig.TokenSource(context.Background(), tokenGoogle)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear servicio de Google"})
		return
	}

	// 3) Obtener datos del usuario (email, nombre, foto, etc.)
	userInfo, err := gservice.NewUserinfoV2MeService(svc).Get().Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener información del usuario"})
		return
	}

	// 4) Guardar o actualizar el usuario en vuestra base de datos
	//    Suponiendo que tienes un modelo Empleado con campos Email, Nombre, Rol, etc.
	var emp models.Empleado
	result := config.DB.Where("email = ?", userInfo.Email).First(&emp)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 4.a) Si no existe, lo creamos con Rol = "empleado" por defecto
			log.Println("Usuario no encontrado, creando nuevo registro:", userInfo)
			emp = models.Empleado{
				Nombre:            userInfo.Name,
				Email:             userInfo.Email,
				Cargo:             "", // o lo que quieras asignar por defecto
				FechaContratacion: time.Now().Format("2006-01-02"),
				SupervisorID:      nil,
				Rol:               "empleado", // nuevo campo
			}
			if err := config.DB.Create(&emp).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear usuario en BD"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar usuario en BD"})
			return
		}
	} else {
		// 4.b) Si ya existe, puedes opcionalmente actualizar Nombre/foto/etc. si lo deseas.
		// emp.Nombre = userInfo.Name
		// config.DB.Save(&emp)
	}

	// 5) Crear las claims del JWT, incluyendo el campo "rol"
	// … tras crear o recuperar emp de la BD …
claims := jwt.MapClaims{
    "sub": emp.ID,
    "rol": emp.Rol,
    "exp": time.Now().Add(time.Hour * 24).Unix(),
}
jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar token JWT"})
    return
}

c.SetCookie("token", tokenString, 3600, "/", "", false, true)
// Hasta aquí guardas el JWT en la cookie “token”
c.Redirect(http.StatusFound, "http://localhost:3001/dashboard")

	
}
