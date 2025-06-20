package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// --------------------------------------------------------------------------------------------------------------------
// HoursWorkedByPeriod returns hours worked over a range, respecting roles:
// - Supervisor: sees their own and their employees’ hours.
//   Optionally filters by empleado_id (if allowed).
// - Employee: sees only their own hours.
func HoursWorkedByPeriod(c *gin.Context) {
	// 1) Read and validate "start" and "end" (YYYY-MM-DD).
	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must specify 'start' and 'end' in YYYY-MM-DD format"})
		return
	}
	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Must be YYYY-MM-DD"})
		return
	}
	// include entire end day
	end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// 2) Define output struct with minutes and hours
	type Result struct {
		EmployeeID    uint    `json:"employee_id"`
		Name          string  `json:"name"`
		TotalHours    float64 `json:"total_hours"`
		TotalMinutes  float64 `json:"total_minutes"`
	}
	var results []Result

	// 3) Extract user_id & role from context
	idRaw, okID := c.Get("usuario_id")
	roleRaw, okRole := c.Get("rol_usuario")
	if !okID || !okRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, castOK := idRaw.(uint)
	role, castOK2 := roleRaw.(string)
	if !castOK || !castOK2 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading user context"})
		return
	}

	// 4) Base query: sum hours and minutes
	db := config.DB.
		Table("fichajes AS f").
		Select(`
			f.empleado_id,
			e.nombre AS name,
			ROUND(SUM(EXTRACT(EPOCH FROM (f.salida - f.entrada)) / 3600)::numeric, 3) AS total_hours,
			ROUND(SUM(EXTRACT(EPOCH FROM (f.salida - f.entrada)) / 60)::numeric, 2) AS total_minutes
		`).
		Joins("JOIN empleados e ON e.id = f.empleado_id").
		Where("f.entrada >= ? AND f.entrada <= ? AND f.salida IS NOT NULL", start, end)

	// 5) Role-based filters
	switch role {
	case "supervisor":
		if empStr := c.Query("employee_id"); empStr != "" {
			empID64, err := strconv.ParseUint(empStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id must be a number"})
				return
			}
			empID := uint(empID64)
			// check belongs to supervisor
			var tmp models.Empleado
			err = config.DB.
				Where("id = ? AND (supervisor_id = ? OR id = ?)", empID, userID, userID).
				First(&tmp).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusForbidden, gin.H{"error": "Cannot view this employee's hours"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying employee"})
				}
				return
			}
			db = db.Where("f.empleado_id = ?", empID)
		} else {
			db = db.Where("e.supervisor_id = ? OR f.empleado_id = ?", userID, userID)
		}

	case "empleado":
		if empStr := c.Query("employee_id"); empStr != "" {
			empID64, err := strconv.ParseUint(empStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id must be a number"})
				return
			}
			empID := uint(empID64)
			if empID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot view other employees' hours"})
				return
			}
			db = db.Where("f.empleado_id = ?", userID)
		} else {
			db = db.Where("f.empleado_id = ?", userID)
		}

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to view hours"})
		return
	}

	// 6) Execute grouped query
	if err := db.
		Group("f.empleado_id, e.nombre").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating hours"})
		return
	}

	// 7) Return JSON
	c.JSON(http.StatusOK, results)
}

// --------------------------------------------------------------------------------------------------------------------
// PunchesByDay (supervisors only) returns count of open/closed punches on a date.
func PunchesByDay(c *gin.Context) {
	idRaw, okID := c.Get("usuario_id")
	roleRaw, okRole := c.Get("rol_usuario")
	if !okID || !okRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, castOK := idRaw.(uint)
	role, castOK2 := roleRaw.(string)
	if !castOK || !castOK2 || role != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: supervisors only"})
		return
	}

	dayStr := c.Query("day")
	if dayStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide 'day' in YYYY-MM-DD format"})
		return
	}
	if _, err := time.Parse("2006-01-02", dayStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'day' format. Must be YYYY-MM-DD"})
		return
	}

	var openCount, closedCount int64
	if err := config.DB.
		Model(&models.Fichaje{}).
		Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
		Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?) AND DATE(fichajes.entrada) = ? AND fichajes.salida IS NULL",
			userID, userID, dayStr).
		Count(&openCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting open punches"})
		return
	}

	if err := config.DB.
		Model(&models.Fichaje{}).
		Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
		Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?) AND DATE(fichajes.salida) = ?",
			userID, userID, dayStr).
		Count(&closedCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting closed punches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"day":             dayStr,
		"punches_open":    openCount,
		"punches_closed":  closedCount,
	})
}

// --------------------------------------------------------------------------------------------------------------------
// VacationsByState (supervisors only) returns counts grouped by state.
func VacationsByState(c *gin.Context) {
	idRaw, okID := c.Get("usuario_id")
	roleRaw, okRole := c.Get("rol_usuario")
	if !okID || !okRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, castOK := idRaw.(uint)
	role, castOK2 := roleRaw.(string)
	if !castOK || !castOK2 || role != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: supervisors only"})
		return
	}

	type StateResult struct {
		State    string `json:"state"`
		Quantity int64  `json:"quantity"`
	}
	var results []StateResult

	if err := config.DB.
		Table("vacacions v").
		Select("v.estado AS state, COUNT(*) AS quantity").
		Joins("JOIN empleados e ON e.id = v.empleado_id").
		Where("e.supervisor_id = ? OR v.empleado_id = ?", userID, userID).
		Group("v.estado").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting vacations by state"})
		return
	}

	c.JSON(http.StatusOK, results)
}
