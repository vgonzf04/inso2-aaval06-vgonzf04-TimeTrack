package models

import "time"

// TimeEntry represents a clock-in/out record for an employee.
// JSON tags use camelCase and English field names.
type TimeEntry struct {
    ID         uint       `gorm:"primaryKey" json:"id"`
    EmployeeID uint       `json:"employee_id"`

    // Preload relation
    Employee   Employee   `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`

    // Raw timestamps (not exposed directly)
    ClockIn    time.Time  `json:"-"`
    ClockOut   *time.Time `json:"-"`

    // Formatted timestamps for JSON
    ClockInStr  string     `gorm:"-" json:"clock_in"`
    ClockOutStr string     `gorm:"-" json:"clock_out,omitempty"`

    Latitude   float64    `json:"latitude"`
    Longitude  float64    `json:"longitude"`
    Location   string     `json:"location"`
}

// FormatTimestamps fills ClockInStr and ClockOutStr using Europe/Madrid timezone.
func (t *TimeEntry) FormatTimestamps() {
    loc, err := time.LoadLocation("Europe/Madrid")
    if err != nil {
        loc = time.UTC
    }
    t.ClockInStr = t.ClockIn.In(loc).Format("2006-01-02 15:04:05")
    if t.ClockOut != nil {
        t.ClockOutStr = t.ClockOut.In(loc).Format("2006-01-02 15:04:05")
    }
}