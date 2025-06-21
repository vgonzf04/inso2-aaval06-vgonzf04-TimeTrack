package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateVacation creates a vacation request.
// Supervisors’ requests auto-approve; employees’ remain “pending”.
func CreateVacation(c *gin.Context) {
	// 1) Bind input JSON
	var input struct {
		StartDate string `json:"startDate" binding:"required"`
		EndDate   string `json:"endDate"   binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing fields"})
		return
	}

	// 2) Parse dates "YYYY-MM-DD"
	start, err1 := time.Parse("2006-01-02", input.StartDate)
	end, err2 := time.Parse("2006-01-02", input.EndDate)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// 3) Validate range
	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "`endDate` must be same or after `startDate`"})
		return
	}

	// 4) Extract auth context
	userIDRaw, hasID := c.Get("user_id")
	roleRaw, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := userIDRaw.(uint)
	userRole, okRole := roleRaw.(string)
	if !okID || !okRole {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading context"})
		return
	}

	// 5) Ensure employee exists
	var emp models.Employee
	if err := config.DB.First(&emp, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching employee"})
		}
		return
	}

	// 6) Determine initial status
	status := "pending"
	if userRole == "supervisor" {
		status = "approved"
	}

	// 7) Create the Vacation record
	vac := models.Vacation{
		EmployeeID: userID,
		Employee:   emp,
		StartDate:  start,
		EndDate:    &end,
		Status:     status,
	}
	if err := config.DB.Create(&vac).Error; err != nil {
		log.Printf("Error creating vacation: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vacation request"})
		return
	}

	c.JSON(http.StatusCreated, vac)
}

// ListVacations returns only the caller’s own vacation requests.
func ListVacations(c *gin.Context) {
	// Auth context
	userIDRaw, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, castOK := userIDRaw.(uint)
	if !castOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading user ID"})
		return
	}

	// Fetch own vacations
	var vacs []models.Vacation
	if err := config.DB.
		Preload("Employee").
		Where("employee_id = ?", userID).
		Find(&vacs).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching your vacation requests"})
		return
	}

	// Format dates for JSON
	for i := range vacs {
		vacs[i].FormatDates()
	}

	c.JSON(http.StatusOK, vacs)
}

// ListEmployeeVacations returns all subordinates’ vacation requests (supervisor only).
func ListEmployeeVacations(c *gin.Context) {
	// Auth + role check
	userIDRaw, hasID := c.Get("user_id")
	roleRaw, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := userIDRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can view employees’ vacations"})
		return
	}

	// Fetch subordinates’ requests
	var vacs []models.Vacation
	if err := config.DB.
		Preload("Employee").
		Joins("JOIN employees e ON e.id = vacations.employee_id").
		Where("e.supervisor_id = ?", userID).
		Find(&vacs).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employees’ vacations"})
		return
	}

	for i := range vacs {
		vacs[i].FormatDates()
	}

	c.JSON(http.StatusOK, vacs)
}

// ApproveVacation allows a supervisor to approve a pending request.
func ApproveVacation(c *gin.Context) {
	// Auth + role
	userIDRaw, hasID := c.Get("user_id")
	roleRaw, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := userIDRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can approve requests"})
		return
	}

	// Vacation ID
	idParam := c.Param("id")
	vacID64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vacation ID"})
		return
	}
	vacID := uint(vacID64)

	// Load and verify belongs to subordinate
	var vac models.Vacation
	if err := config.DB.
		Joins("JOIN employees e ON e.id = vacations.employee_id").
		Where("vacations.id = ? AND e.supervisor_id = ?", vacID, userID).
		First(&vac).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found or not your subordinate’s"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching request"})
		}
		return
	}

	// Only pending can be approved
	if vac.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be approved"})
		return
	}

	// Approve and save
	vac.Status = "approved"
	if err := config.DB.Save(&vac).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
		return
	}

	config.DB.Preload("Employee").First(&vac, vac.ID)
	c.JSON(http.StatusOK, vac)
}

// RejectVacation allows a supervisor to reject a pending request.
func RejectVacation(c *gin.Context) {
	// Auth + role
	userIDRaw, hasID := c.Get("user_id")
	roleRaw, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := userIDRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can reject requests"})
		return
	}

	// Vacation ID
	idParam := c.Param("id")
	vacID64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vacation ID"})
		return
	}
	vacID := uint(vacID64)

	// Load and verify subordinate
	var vac models.Vacation
	if err := config.DB.
		Joins("JOIN employees e ON e.id = vacations.employee_id").
		Where("vacations.id = ? AND e.supervisor_id = ?", vacID, userID).
		First(&vac).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found or not your subordinate’s"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching request"})
		}
		return
	}

	// Only pending can be rejected
	if vac.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be rejected"})
		return
	}

	// Reject and save
	vac.Status = "rejected"
	if err := config.DB.Save(&vac).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
		return
	}

	config.DB.Preload("Employee").First(&vac, vac.ID)
	c.JSON(http.StatusOK, vac)
}
