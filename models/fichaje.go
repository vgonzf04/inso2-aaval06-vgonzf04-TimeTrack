package models

import "time"

type Fichaje struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	EmpleadoID uint       `json:"empleado_id"`
	Entrada    time.Time  `json:"entrada"`
	Salida     *time.Time `json:"salida"`
	Ubicacion  string     `json:"ubicacion"`
}
