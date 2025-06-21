package models

import "time"

// TimeEntry represents a clock-in/out record for an employee.
// JSON tags use camelCase and English field names to match the controllers.
type TimeEntry struct {
    ID         uint       `gorm:"primaryKey" json:"id"`
    EmployeeID uint       `json:"employee_id"`

    // Preload relation
    Employee   Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

    // Raw timestamps (not exposed directly)
    StartTime  time.Time  `json:"-"`
    EndTime    *time.Time `json:"-"`

    // Formatted timestamps for JSON
    Start      string     `gorm:"-" json:"start"`
    End        string     `gorm:"-" json:"end,omitempty"`

    Latitude   float64    `json:"lat"`
    Longitude  float64    `json:"lng"`
    Location   string     `json:"location"`
}

// FormatDates fills Start and End using Europe/Madrid timezone.
func (t *TimeEntry) FormatDates() {
    loc, err := time.LoadLocation("Europe/Madrid")
    if err != nil {
        loc = time.UTC
    }
    t.Start = t.StartTime.In(loc).Format("2006-01-02 15:04:05")
    if t.EndTime != nil {
        t.End = t.EndTime.In(loc).Format("2006-01-02 15:04:05")
    }
}