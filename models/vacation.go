package models

import "time"

type Vacacion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmpleadoID uint      `json:"empleado_id"`
	Inicio     time.Time `json:"inicio"`
	Fin        time.Time `json:"fin"`
	Estado     string    `json:"estado"` // "pendiente", "aprobada", "rechazada"
}
