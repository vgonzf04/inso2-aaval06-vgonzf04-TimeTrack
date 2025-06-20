package models

import "time"

// Vacation represents a time-off request by an employee.
type Vacation struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	EmployeeID   uint      `gorm:"not null" json:"employee_id"`   // ID of the requesting employee
	Employee     Employee  `gorm:"foreignKey:EmployeeID" json:"employee"`

	// Actual date fields
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`                     // nil until end date is set

	Status       string     `json:"status"`                       // "pending", "approved", "rejected"

	// Auxiliary fields for formatted dates (DD-MM-YYYY)
	StartDateStr string     `gorm:"-" json:"start_date_str"`
	EndDateStr   string     `gorm:"-" json:"end_date_str"`
}

// FormatDates fills StartDateStr and EndDateStr in "02-01-2006" format.
func (v *Vacation) FormatDates() {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}

	v.StartDateStr = v.StartDate.In(loc).Format("02-01-2006")
	if v.EndDate != nil {
		v.EndDateStr = v.EndDate.In(loc).Format("02-01-2006")
	} else {
		v.EndDateStr = ""
	}
}
