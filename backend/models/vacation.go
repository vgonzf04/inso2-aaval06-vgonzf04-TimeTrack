package models

import "time"

type Vacacion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmpleadoID uint      `gorm:"not null" json:"empleado_id"` // ID del empleado que solicita la vacación
	Empleado   Empleado  `gorm:"foreignKey:EmpleadoID" json:"empleado"`

	Inicio time.Time  `json:"inicio"` // se serializará como RFC3339
	Fin    *time.Time `json:"fin"`    // puntero para poder ser nil si aún no se ha establecido

	Estado string `json:"estado"` // "pendiente", "aprobada", "rechazada"`

	// Campos auxiliares (no se guardan en BD) para enviar la fecha formateada "dd-mm-aaaa"
	InicioStr string `gorm:"-" json:"inicioStr"`
	FinStr    string `gorm:"-" json:"finStr"`
}

// FormatearFechas rellena InicioStr y FinStr en formato "dd-mm-aaaa"
func (v *Vacacion) FormatearFechas() {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}

	// Formatear Inicio como "02-01-2006"
	v.InicioStr = v.Inicio.In(loc).Format("02-01-2006")

	// Si Fin no es nil, formatearlo; si es nil, dejar cadena vacía
	if v.Fin != nil {
		v.FinStr = v.Fin.In(loc).Format("02-01-2006")
	} else {
		v.FinStr = ""
	}
}
