package models

import "time"

type Vacacion struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    EmpleadoID uint      `json:"empleado_id"`
    Empleado   Empleado  `gorm:"foreignKey:EmpleadoID" json:"empleado,omitempty"`

    Inicio     time.Time `json:"inicio"` // se serializará como RFC3339, pero al parsear "YYYY-MM-DD" funciona
    Fin        time.Time `json:"fin"`
    Estado     string    `json:"estado"` // "pendiente", "aprobada", "rechazada"
}

