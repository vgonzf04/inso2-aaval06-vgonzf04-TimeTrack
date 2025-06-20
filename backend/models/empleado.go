package models

// Employee represents a company worker
type Employee struct {
    ID uint `gorm:"primaryKey" json:"id"` // Unique auto-incrementing ID

    Name     string  `json:"name"`               // Employee’s full name
    Email    string  `gorm:"unique" json:"email"`// Unique email (used for login/ID)
    Position string  `json:"position"`           // Job title (e.g. Developer, HR, etc.)

    HireDate string  `json:"hire_date"`          // ISO date YYYY-MM-DD when they were hired

    // SupervisorID is optional: nil if no direct supervisor assigned
    SupervisorID *uint  `json:"supervisor_id"`    // ID of their supervisor (another employee)
    Role         string `json:"role"`             // "employee" or "supervisor"
}
