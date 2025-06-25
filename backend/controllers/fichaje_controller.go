package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateTimeEntry marks the start of a time entry.
// Expects JSON: { "lat": <float>, "lng": <float> }
func CreateTimeEntry(c *gin.Context) {
	// 1) Authentication context
	idRaw, existsID := c.Get("user_id")
	roleRaw, existsRol := c.Get("user_role")
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
	if userRole != "employee" && userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to create time entries"})
		return
	}

	// 2) Verify employee exists
	var emp models.Employee
	if err := config.DB.First(&emp, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			log.Printf("Error fetching employee: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching employee"})
		}
		return
	}

	// 3) Bind JSON input
	var input struct {
		Latitude float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("JSON bind error in CreateTimeEntry: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4) Ensure no open time entry exists
	var open models.TimeEntry
	err := config.DB.
		Where("employee_id = ? AND end_time IS NULL", emp.ID).
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

	// 5) Reverse geocode
	address := reverseGeocode(input.Latitude, input.Longitude)

	// 6) Create the new time entry
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("Could not load Europe/Madrid, using UTC: %v\n", err)
		loc = time.UTC
	}
	now := time.Now().In(loc)

	entry := models.TimeEntry{
		EmployeeID: emp.ID,
		StartTime:  now,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		Location:   address,
	}

	if err := config.DB.Create(&entry).Error; err != nil {
		log.Printf("Error creating time entry: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create time entry"})
		return
	}

	// 7) Preload employee relationship
	if err := config.DB.Preload("Employee").First(&entry, entry.ID).Error; err != nil {
		log.Printf("Preload error: %v\n", err)
	}

	// 8) Format the timestamps
	entry.FormatDates()

	// 9) Return the created entry
	c.JSON(http.StatusCreated, entry)
}

// CloseTimeEntry marks the end of an open time entry.
func CloseTimeEntry(c *gin.Context) {
	// 1) Parse entry ID from URL
	id := c.Param("id")

	// 2) Load the entry (with employee)
	var entry models.TimeEntry
	if err := config.DB.Preload("Employee").First(&entry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Time entry not found"})
		} else {
			log.Printf("Error loading entry ID=%s: %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error loading time entry"})
		}
		return
	}

	// 3) Check it's not already closed
	if entry.EndTime != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This time entry is already closed"})
		return
	}

	// 4) Set end time
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("Could not load Europe/Madrid, using UTC: %v\n", err)
		loc = time.UTC
	}
	exitTime := time.Now().In(loc)

	// 5) Update only the EndTime field
	if err := config.DB.Model(&entry).
		Update("end_time", &exitTime).
		Error; err != nil {
		log.Printf("Error closing entry ID=%s: %v\n", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close time entry"})
		return
	}

	// 6) Reload with employee
	if err := config.DB.Preload("Employee").First(&entry, id).Error; err != nil {
		log.Printf("Reload error after closing entry ID=%s: %v\n", id, err)
	}

	// 7) Format the timestamps
	entry.FormatDates()

	// 8) Return the closed entry
	c.JSON(http.StatusOK, entry)
}

// GetCurrentTimeEntry returns the currently open time entry for the user.
func GetCurrentTimeEntry(c *gin.Context) {
	idRaw, existsID := c.Get("user_id")
	roleRaw, existsRol := c.Get("user_role")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := idRaw.(uint)
	userRole, _ := roleRaw.(string)

	var entry models.TimeEntry
	db := config.DB.Preload("Employee")

	switch userRole {
	case "supervisor":
		err := db.
			Joins("JOIN employees e ON e.id = time_entries.employee_id").
			Where("(e.supervisor_id = ? OR time_entries.employee_id = ?) AND time_entries.end_time IS NULL", userID, userID).
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

	case "employee":
		err := db.
			Where("employee_id = ? AND end_time IS NULL", userID).
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

// ListTimeEntries returns historic entries for the user.
func ListTimeEntries(c *gin.Context) {
	idRaw, existsID := c.Get("user_id")
	roleRaw, existsRol := c.Get("user_role")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := idRaw.(uint)
	userRole, _ := roleRaw.(string)

	var entries []models.TimeEntry
	db := config.DB.Preload("Employee")

	switch userRole {
	case "supervisor":
		if err := db.
			Joins("JOIN employees e ON e.id = time_entries.employee_id").
			Where("e.supervisor_id = ? OR time_entries.employee_id = ?", userID, userID).
			Find(&entries).Error; err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error listing time entries"})
			return
		}

	case "employee":
		if err := db.
			Where("employee_id = ?", userID).
			Find(&entries).Error; err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error listing time entries"})
			return
		}

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to list time entries"})
		return
	}

	// Format all the timestamps
	for i := range entries {
		entries[i].FormatDates()
	}

	c.JSON(http.StatusOK, entries)
}

// reverseGeocode calls Google Maps API and returns a formatted address.
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
