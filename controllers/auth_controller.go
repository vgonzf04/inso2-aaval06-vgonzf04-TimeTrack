package controllers

import (
	"context"
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
            emp = models.Empleado{
                Nombre:            userInfo.Name,
                Email:             userInfo.Email,
                Cargo:             "",      // o lo que quieras asignar por defecto
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
    claims := jwt.MapClaims{
        "sub": emp.ID,         // ID del empleado en BD
        "rol": emp.Rol,        // rol: "empleado" o "supervisor"
        "exp": time.Now().Add(time.Hour * 24).Unix(), // expira en 24h
    }
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    // 6) Firmar el token con tu JWT_SECRET (defínelo en .env)
    tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar token JWT"})
        return
    }

    // 7) Devolver el token JWT junto con la info básica (opcional)
    c.JSON(http.StatusOK, gin.H{
        "token": tokenString,
        "usuario": gin.H{
            "id":    emp.ID,
            "nombre": emp.Nombre,
            "email":  emp.Email,
            "rol":    emp.Rol,
        },
    })
}

