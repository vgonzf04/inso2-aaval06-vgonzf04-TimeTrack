package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// CreateTimeEntry marks the start of a time entry.
// Expects JSON: { "lat": <float>, "lng": <float> }
func CreateTimeEntry(c *gin.Context) {
	// 1) Authentication context
	idRaw, existsID := c.Get("usuario_id")
	roleRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := idRaw.(uint)
	userRole, okRole := roleRaw.(string)
	if !okID || !okRole {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading context"})
		return
	}
	// Only employees or supervisors may punch in
	if userRole != "empleado" && userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to create time entries"})
		return
	}

	// Verify employee exists
	var emp models.Empleado
	if err := config.DB.First(&emp, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			log.Printf("Error fetching employee: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching employee"})
		}
		return
	}

	// 2) Bind JSON input
	var input struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("JSON bind error in CreateTimeEntry: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3) Ensure no open entry exists
	var open models.Fichaje
	err := config.DB.
		Where("empleado_id = ? AND salida IS NULL", emp.ID).
		First(&open).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an open time entry. Close it first"})
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Error checking open entries: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking open entries"})
		return
	}

	// 4) Reverse geocode
	address := reverseGeocode(input.Lat, input.Lng)

	// 5) Create entry
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("Could not load Europe/Madrid, using UTC: %v\n", err)
		loc = time.UTC
	}
	now := time.Now().In(loc)

	entry := models.Fichaje{
		EmpleadoID: emp.ID,
		Entrada:    now,
		Latitud:    input.Lat,
		Longitud:   input.Lng,
		Ubicacion:  address,
	}

	if err := config.DB.Create(&entry).Error; err != nil {
		log.Printf("Error creating time entry: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create time entry"})
		return
	}

	// 6) Preload employee
	if err := config.DB.Preload("Empleado").First(&entry, entry.ID).Error; err != nil {
		log.Printf("Preload error: %v\n", err)
	}

	// 7) Format dates
	entry.FormatearFechas()

	// 8) Return
	c.JSON(http.StatusCreated, entry)
}

// CloseTimeEntry marks the end of an open time entry.
// No body required.
func CloseTimeEntry(c *gin.Context) {
	// 1) ID from URL
	id := c.Param("id")

	// 2) Load entry
	var entry models.Fichaje
	if err := config.DB.Preload("Empleado").First(&entry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		} else {
			log.Printf("Error loading entry ID=%s: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error loading time entry"})
		}
		return
	}

	// 3) Already closed?
	if entry.Salida != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This time entry is already closed"})
		return
	}

	// 4) Set exit time
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("Could not load Europe/Madrid, using UTC: %v\n", err)
		loc = time.UTC
	}
	exitTime := time.Now().In(loc)

	// 5) Update only salida
	if err := config.DB.Model(&entry).Updates(map[string]interface{}{"salida": &exitTime}).Error; err != nil {
		log.Printf("Error updating exit time for ID=%s: %v\n", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close time entry"})
		return
	}

	// 6) Reload
	if err := config.DB.Preload("Empleado").First(&entry, id).Error; err != nil {
		log.Printf("Reload error after closing entry ID=%s: %v\n", id, err)
	}

	// 7) Format dates
	entry.FormatearFechas()

	// 8) Return
	c.JSON(http.StatusOK, entry)
}

// GetCurrentTimeEntry returns the currently open time entry for the user.
func GetCurrentTimeEntry(c *gin.Context) {
	// 1) Auth context
	idRaw, existsID := c.Get("usuario_id")
	roleRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := idRaw.(uint)
	userRole, okRole := roleRaw.(string)
	if !okID || !okRole {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading context"})
		return
	}

	// 2) Query open entry
	var entry models.Fichaje
	db := config.DB.Preload("Empleado")

	switch userRole {
	case "supervisor":
		err := db.
			Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
			Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?) AND fichajes.salida IS NULL", userID, userID).
			First(&entry).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "No open time entry for this user"})
			} else {
				log.Printf("Error fetching current entry: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching current time entry"})
			}
			return
		}
	case "empleado":
		err := db.
			Where("empleado_id = ? AND salida IS NULL", userID).
			First(&entry).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "No open time entry for this user"})
			} else {
				log.Printf("Error fetching current entry: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching current time entry"})
			}
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to view current time entry"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// ListTimeEntries returns historic entries.
// Supports optional ?empleado_id, ?from=YYYY-MM-DD, ?to=YYYY-MM-DD
func ListTimeEntries(c *gin.Context) {
	// 1) Auth context
	idRaw, existsID := c.Get("usuario_id")
	roleRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := idRaw.(uint)
	userRole, okRole := roleRaw.(string)
	if !okID || !okRole {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading context"})
		return
	}

	// 2) Build query
	var entries []models.Fichaje
	db := config.DB.Preload("Empleado")

	switch userRole {
	case "supervisor":
		if err := db.
			Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
			Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?)", userID, userID).
			Find(&entries).Error; err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error listing time entries"})
			return
		}

	case "empleado":
		if err := db.
			Where("empleado_id = ?", userID).
			Find(&entries).Error; err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error listing time entries"})
			return
		}

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to list time entries"})
		return
	}

	// 3) Format dates
	for i := range entries {
		entries[i].FormatearFechas()
	}

	c.JSON(http.StatusOK, entries)
}

// reverseGeocode calls Google Maps API Reverse Geocoding
func reverseGeocode(lat, lng float64) string {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		return ""
	}
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=%s",
		lat, lng, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error calling Google Maps: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Google Maps body: %v\n", err)
		return ""
	}

	var result struct {
		Results []struct {
			FormattedAddress string   `json:"formatted_address"`
			Types            []string `json:"types"`
		} `json:"results"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing Google Maps JSON: %v\n", err)
		return ""
	}
	if result.Status != "OK" || len(result.Results) == 0 {
		return ""
	}

	for _, r := range result.Results {
		for _, t := range r.Types {
			if t == "street_address" || t == "route" {
				return r.FormattedAddress
			}
		}
	}
	return result.Results[0].FormattedAddress
}
