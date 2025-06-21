package controllers

import (
	"net/http"
	"strconv"
	"time"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HoursWorkedByPeriod returns hours and minutes worked over a range, respecting roles:
// - Supervisor: sees own and subordinates’ entries (optional filter by employee_id).
// - Employee: sees only own entries.
func HoursWorkedByPeriod(c *gin.Context) {
	// 1) Parse and validate "start" / "end"
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

	// 2) Output struct
	type Result struct {
		EmployeeID   uint    `json:"employee_id"`
		Name         string  `json:"name"`
		TotalHours   float64 `json:"total_hours"`
		TotalMinutes float64 `json:"total_minutes"`
	}
	var results []Result

	// 3) Extract user_id & role
	rawID, hasID := c.Get("user_id")
	rawRole, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := rawID.(uint)
	role, okRole := rawRole.(string)
	if !okID || !okRole {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error reading user context"})
		return
	}

	// 4) Base query on time_entries
	db := config.DB.
		Table("time_entries AS t").
		Select(`
			t.employee_id,
			e.name,
			ROUND(SUM(EXTRACT(EPOCH FROM (t.end_time - t.start_time)) / 3600)::numeric, 3)   AS total_hours,
			ROUND(SUM(EXTRACT(EPOCH FROM (t.end_time - t.start_time)) / 60)::numeric, 2)    AS total_minutes
		`).
		Joins("JOIN employees e ON e.id = t.employee_id").
		Where("t.start_time >= ? AND t.start_time <= ? AND t.end_time IS NOT NULL", start, end)

	// 5) Role-based filtering
	switch role {
	case "supervisor":
		if empStr := c.Query("employee_id"); empStr != "" {
			empID64, err := strconv.ParseUint(empStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id must be a number"})
				return
			}
			empID := uint(empID64)
			// verify subordinate or self
			var tmp models.Employee
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
			db = db.Where("t.employee_id = ?", empID)
		} else {
			db = db.Where("e.supervisor_id = ? OR t.employee_id = ?", userID, userID)
		}

	case "employee":
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
			db = db.Where("t.employee_id = ?", userID)
		} else {
			db = db.Where("t.employee_id = ?", userID)
		}

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Role not authorized to view hours"})
		return
	}

	// 6) Execute grouped query
	if err := db.
		Group("t.employee_id, e.name").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error calculating hours"})
		return
	}

	// 7) Return results
	c.JSON(http.StatusOK, results)
}

// PunchesByDay (supervisors only) returns counts of open/closed punches for a date.
func PunchesByDay(c *gin.Context) {
	rawID, hasID := c.Get("user_id")
	rawRole, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := rawID.(uint)
	role, okRole := rawRole.(string)
	if !okID || !okRole || role != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: supervisors only"})
		return
	}

	day := c.Query("day")
	if day == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must provide 'day' in YYYY-MM-DD format"})
		return
	}
	if _, err := time.Parse("2006-01-02", day); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'day' format. Must be YYYY-MM-DD"})
		return
	}

	var openCount, closedCount int64
	// open punches
	if err := config.DB.
		Model(&models.TimeEntry{}).
		Joins("JOIN employees e ON e.id = time_entries.employee_id").
		Where("(e.supervisor_id = ? OR time_entries.employee_id = ?) AND DATE(time_entries.start_time) = ? AND time_entries.end_time IS NULL",
			userID, userID, day).
		Count(&openCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting open punches"})
		return
	}
	// closed punches
	if err := config.DB.
		Model(&models.TimeEntry{}).
		Joins("JOIN employees e ON e.id = time_entries.employee_id").
		Where("(e.supervisor_id = ? OR time_entries.employee_id = ?) AND DATE(time_entries.end_time) = ?",
			userID, userID, day).
		Count(&closedCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting closed punches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"day":            day,
		"punches_open":   openCount,
		"punches_closed": closedCount,
	})
}

// VacationsByState (supervisors only) returns counts grouped by state.
func VacationsByState(c *gin.Context) {
	rawID, hasID := c.Get("user_id")
	rawRole, hasRole := c.Get("user_role")
	if !hasID || !hasRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	userID, okID := rawID.(uint)
	role, okRole := rawRole.(string)
	if !okID || !okRole || role != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: supervisors only"})
		return
	}

	type StateResult struct {
		State    string `json:"state"`
		Quantity int64  `json:"quantity"`
	}
	var results []StateResult

	if err := config.DB.
		Table("vacations v").
		Select("v.status AS state, COUNT(*) AS quantity").
		Joins("JOIN employees e ON e.id = v.employee_id").
		Where("e.supervisor_id = ? OR v.employee_id = ?", userID, userID).
		Group("v.status").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting vacations by state"})
		return
	}

	c.JSON(http.StatusOK, results)
}
