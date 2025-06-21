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

// Me returns the current user's role based on the "token" cookie.
func Me(c *gin.Context) {
	// 1) Read the “token” cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// 2) Parse and validate the JWT with the same secret
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 3) Extract claims and get “role”
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading token"})
		return
	}
	roleVal, ok := claims["role"]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Role claim not found"})
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role format in token"})
		return
	}

	// 4) Return JSON with the role
	c.JSON(http.StatusOK, gin.H{"role": role})
}

// GoogleLogin redirects the user to Google's OAuth2 consent page.
func GoogleLogin(c *gin.Context) {
	url := config.GoogleOAuthConfig.AuthCodeURL(
		"random-state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth2 callback from Google.
func GoogleCallback(c *gin.Context) {
	code := c.Query("code")

	// 1) Exchange the code for a Google token
	tokenGoogle, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token with Google"})
		return
	}

	// 2) Create Google service to fetch user info
	svc, err := gservice.NewService(
		context.Background(),
		option.WithTokenSource(config.GoogleOAuthConfig.TokenSource(context.Background(), tokenGoogle)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google service"})
		return
	}

	// 3) Get user info (email, name, picture, etc.)
	userInfo, err := gservice.NewUserinfoV2MeService(svc).Get().Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info from Google"})
		return
	}

	// 4) Save or update the user in the database
	var emp models.Employee
	result := config.DB.Where("email = ?", userInfo.Email).First(&emp)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 4.a) If not found, create with default role "employee"
			log.Println("User not found, creating new record:", userInfo)
			emp = models.Employee{
				Name:       userInfo.Name,
				Email:      userInfo.Email,
				Position:   "",
				HireDate:   time.Now().Format("2006-01-02"),
				SupervisorID: nil,
				Role:       "employee",
			}
			if err := config.DB.Create(&emp).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user in database"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database lookup error"})
			return
		}
	}

	// 5) Build JWT claims including "role"
	claims := jwt.MapClaims{
		"sub":  emp.ID,
		"role": emp.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	// Set the cookie and redirect to the dashboard
	http.SetCookie(c.Writer, &http.Cookie{
        Name:     "token",
        Value:    tokenString,
        Path:     "/",
        MaxAge:   3600,                      // en segundos (1 hora). También podrías usar Expires: time.Now().Add(...)
        HttpOnly: true,                      // ← no accesible desde JS
        Secure:   false,                      // ← sólo por HTTPS
        SameSite: http.SameSiteLaxMode,     // ← permite el envío cross‐site
    })
	c.Redirect(http.StatusFound, "http://localhost:3001/dashboard")
}

// Logout clears the authentication cookie.
func Logout(c *gin.Context) {
	// To delete a cookie, set MaxAge < 0
	http.SetCookie(c.Writer, &http.Cookie{
        Name:     "token",
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
    })
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
