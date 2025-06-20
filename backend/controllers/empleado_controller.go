package controllers

import (
	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetEmployeeByID returns an employee by its ID.
// Only supervisors may call this endpoint.
func GetEmployeeByID(c *gin.Context) {
	// 1) Retrieve user_id and user_role from context (set by JWTAuth)
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

	// 2) Only supervisors can fetch arbitrary employees
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: only supervisors can query employees by ID"})
		return
	}

	// 3) Parse the requested employee ID from URL
	paramID := c.Param("id")
	empID64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	empID := uint(empID64)

	// 4) Load employee where id = empID AND supervisor_id = userID
	var emp models.Empleado
	if err := config.DB.
		Where("id = ? AND supervisor_id = ?", empID, userID).
		First(&emp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found or not under your supervision"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching employee"})
		}
		return
	}

	// 5) Return the employee
	c.JSON(http.StatusOK, emp)
}

// GetMyProfile returns the profile of the authenticated user.
func GetMyProfile(c *gin.Context) {
	// 1) Extract user_id from context
	idRaw, existsID := c.Get("usuario_id")
	if !existsID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, ok := idRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading context"})
		return
	}

	// 2) Query the employee by their own ID
	var emp models.Empleado
	if err := config.DB.First(&emp, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching profile"})
		}
		return
	}

	// 3) Return the employee record
	c.JSON(http.StatusOK, emp)
}

// CreateEmployee adds a new employee.
func CreateEmployee(c *gin.Context) {
	var input models.Empleado
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Set hiring date to today
	input.FechaContratacion = time.Now().Format("2006-01-02")
	fmt.Printf("✅ Received JSON: %+v\n", input)

	// Normalize and validate role
	role := strings.ToLower(strings.TrimSpace(input.Rol))
	if role == "" || role == "empleado" {
		input.Rol = "empleado"
	} else if role == "supervisor" {
		input.Rol = "supervisor"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role: only 'empleado' or 'supervisor' allowed"})
		return
	}

	// If a supervisor_id was provided, verify it exists and is a supervisor
	if input.SupervisorID != nil {
		var sup models.Empleado
		if err := config.DB.First(&sup, *input.SupervisorID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Specified supervisor does not exist"})
			return
		}
		if strings.ToLower(sup.Rol) != "supervisor" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assigned supervisor must have role 'supervisor'"})
			return
		}
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// UpdateEmployee modifies an existing employee.
// Only supervisors may perform updates, and only on their own subordinates.
func UpdateEmployee(c *gin.Context) {
	// 1) Extract user_id and role
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
	// 2) Only supervisors may update
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: only supervisors may update employees"})
		return
	}

	// Parse target employee ID
	paramID := c.Param("id")
	empID64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	empID := uint(empID64)

	// 3) Load the employee if it's under this supervisor
	var emp models.Empleado
	if err := config.DB.
		Where("id = ? AND supervisor_id = ?", empID, userID).
		First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found or not under your supervision"})
		return
	}

	// 4) Bind update payload
	var payload struct {
		Nombre            string `json:"nombre"`
		Email             string `json:"email"`
		Cargo             string `json:"cargo"`
		FechaContratacion string `json:"fecha_contratacion"`
		SupervisorID      *uint  `json:"supervisor_id"`
		Rol               string `json:"rol"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 5) Apply updates if provided
	if payload.Nombre != "" {
		emp.Nombre = payload.Nombre
	}
	if payload.Email != "" {
		emp.Email = payload.Email
	}
	if payload.Cargo != "" {
		emp.Cargo = payload.Cargo
	}
	if payload.FechaContratacion != "" {
		emp.FechaContratacion = payload.FechaContratacion
	}
	emp.SupervisorID = payload.SupervisorID

	if payload.Rol != "" {
		r := strings.ToLower(strings.TrimSpace(payload.Rol))
		if r != "empleado" && r != "supervisor" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role: only 'empleado' or 'supervisor'"})
			return
		}
		emp.Rol = r
	}

	// 6) Save changes
	if err := config.DB.Save(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving changes"})
		return
	}

	c.JSON(http.StatusOK, emp)
}

// DeleteEmployee removes an employee by ID.
// Only supervisors may delete, and only their own subordinates.
func DeleteEmployee(c *gin.Context) {
	// 1) Extract user_id and role
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
	// 2) Only supervisors may delete
	if userRole != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: only supervisors may delete employees"})
		return
	}

	// Parse target ID
	paramID := c.Param("id")
	empID64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	empID := uint(empID64)

	// 3) Load subordinate
	var emp models.Empleado
	if err := config.DB.
		Where("id = ? AND supervisor_id = ?", empID, userID).
		First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found or not under your supervision"})
		return
	}

	// 4) Delete
	if err := config.DB.Delete(&emp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting employee"})
		return
	}

	// 5) Return no content
	c.Status(http.StatusNoContent)
}

// GetAuthenticatedUser returns basic info about the logged-in user.
func GetAuthenticatedUser(c *gin.Context) {
	// 1) Extract user_id and role
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

	// 2) Load user record
	var emp models.Empleado
	if err := config.DB.First(&emp, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching profile"})
		return
	}

	// 3) Return only non-sensitive fields
	c.JSON(http.StatusOK, gin.H{
		"id":                  emp.ID,
		"name":                emp.Nombre,
		"email":               emp.Email,
		"position":            emp.Cargo,
		"hiring_date":         emp.FechaContratacion,
		"supervisor_id":       emp.SupervisorID,
		"role":                userRole,
	})
}
