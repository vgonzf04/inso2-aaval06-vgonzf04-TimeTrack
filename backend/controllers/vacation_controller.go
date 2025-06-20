package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// CreateVacation creates a vacation request.
// Supervisors’ requests auto-approve; employees’ remain “pending”.
func CreateVacation(c *gin.Context) {
	// 1. Bind input JSON
	var input struct {
		Start string `json:"start" binding:"required"`
		End   string `json:"end"   binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing fields"})
		return
	}

	// 2. Parse dates
	startDate, err1 := time.Parse("2006-01-02", input.Start)
	endDate,   err2 := time.Parse("2006-01-02", input.End)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// 3. Validate range
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "`end` must be same or after `start`"})
		return
	}

	// 4. Auth context
	idRaw, hasID := c.Get("usuario_id")
	roleRaw, hasRole := c.Get("rol_usuario")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userRole, ok := roleRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading user role"})
		return
	}

	// 5. Ensure employee exists
	var emp models.Empleado
	if err := config.DB.First(&emp, idRaw).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching employee"})
		}
		return
	}

	// 6. Determine initial status
	status := "pending"
	if userRole == "supervisor" {
		status = "approved"
	}

	// 7. Create vacation record
	vac := models.Vacacion{
		EmpleadoID: emp.ID,
		Empleado:   emp,
		Inicio:     startDate,
		Fin:        &endDate,
		Estado:     status,
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
	idRaw, ok := c.Get("usuario_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, castOK := idRaw.(uint)
	if !castOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading user ID"})
		return
	}

	// Fetch own vacations
	var vacs []models.Vacacion
	if err := config.DB.Preload("Empleado").
		Where("empleado_id = ?", userID).
		Find(&vacs).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching your vacation requests"})
		return
	}

	// Format dates
	for i := range vacs {
		vacs[i].FormatearFechas()
	}

	c.JSON(http.StatusOK, vacs)
}

// ListEmployeeVacations returns all subordinates’ vacation requests (supervisor only).
func ListEmployeeVacations(c *gin.Context) {
	// Auth + role check
	idRaw, hasID := c.Get("usuario_id")
	roleRaw, hasRole := c.Get("rol_usuario")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := idRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can view employees’ vacations"})
		return
	}

	// Fetch subordinates’ requests
	var vacs []models.Vacacion
	if err := config.DB.
		Preload("Empleado").
		Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
		Where("e.supervisor_id = ?", userID).
		Find(&vacs).Error; err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employees’ vacations"})
		return
	}

	for i := range vacs {
		vacs[i].FormatearFechas()
	}

	c.JSON(http.StatusOK, vacs)
}

// ApproveVacation allows a supervisor to approve a pending request.
func ApproveVacation(c *gin.Context) {
	// Auth + role
	idRaw, hasID := c.Get("usuario_id")
	roleRaw, hasRole := c.Get("rol_usuario")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := idRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can approve requests"})
		return
	}

	// Vacation ID
	param := c.Param("id")
	vacID, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vacation ID"})
		return
	}

	// Load and verify belongs to subordinate
	var vac models.Vacacion
	result := config.DB.
		Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
		Where("vacacions.id = ? AND e.supervisor_id = ?", uint(vacID), userID).
		First(&vac)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found or not your subordinate’s"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching request"})
		}
		return
	}

	// Only pending can be approved
	if vac.Estado != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be approved"})
		return
	}

	// Approve and save
	vac.Estado = "approved"
	if err := config.DB.Save(&vac).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
		return
	}

	config.DB.Preload("Empleado").First(&vac, vac.ID)
	c.JSON(http.StatusOK, vac)
}

// RejectVacation allows a supervisor to reject a pending request.
func RejectVacation(c *gin.Context) {
	// Auth + role
	idRaw, hasID := c.Get("usuario_id")
	roleRaw, hasRole := c.Get("rol_usuario")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, _ := idRaw.(uint)
	userRole, _ := roleRaw.(string)
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only supervisors can reject requests"})
		return
	}

	// Vacation ID
	param := c.Param("id")
	vacID, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vacation ID"})
		return
	}

	// Load and verify subordinate
	var vac models.Vacacion
	result := config.DB.
		Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
		Where("vacacions.id = ? AND e.supervisor_id = ?", uint(vacID), userID).
		First(&vac)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Request not found or not your subordinate’s"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching request"})
		}
		return
	}

	// Only pending can be rejected
	if vac.Estado != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending requests can be rejected"})
		return
	}

	// Reject and save
	vac.Estado = "rejected"
	if err := config.DB.Save(&vac).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
		return
	}

	config.DB.Preload("Empleado").First(&vac, vac.ID)
	c.JSON(http.StatusOK, vac)
}
