package models

// Empleado representa a un trabajador de la empresa
type Empleado struct {
	ID uint `gorm:"primaryKey" json:"id"` // ID único autoincremental

	Nombre string `json:"nombre"`              // Nombre del empleado
	Email  string `gorm:"unique" json:"email"` // Email único (importante para login o identificación)
	Cargo  string `json:"cargo"`               // Puesto o rol (ej: Desarrollador, RRHH, etc.)

	FechaContratacion string `json:"fecha_contratacion"`

	// SupervisorID es opcional: puede ser null si el empleado no tiene jefe directo asignado
	SupervisorID *uint  `json:"supervisor_id"` // ID de su supervisor (otro empleado)
	Rol          string `json:"rol"`
}
