package models

import "time"

type Fichaje struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	EmpleadoID uint       `json:"empleado_id"`

	// Relación para Preload("Empleado"):
	Empleado   Empleado   `gorm:"foreignKey:EmpleadoID" json:"empleado,omitempty"`

	// Fecha real, no expuesta directamente en JSON
	Entrada    time.Time  `json:"-"`
	Salida     *time.Time `json:"-"`

	// Campos auxiliares para JSON
	EntradaStr string     `gorm:"-" json:"entrada"`
	SalidaStr  string     `gorm:"-" json:"salida,omitempty"`

	Latitud    float64    `json:"latitud"`
	Longitud   float64    `json:"longitud"`
	Ubicacion  string     `json:"ubicacion"`
}

// FormatearFechas rellena los campos EntradaStr y SalidaStr
func (f *Fichaje) FormatearFechas() {
	f.EntradaStr = f.Entrada.Format("2006-01-02 15:04:05")
	if f.Salida != nil {
		f.SalidaStr = f.Salida.Format("2006-01-02 15:04:05")
	}
}
