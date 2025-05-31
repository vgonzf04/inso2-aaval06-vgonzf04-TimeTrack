package models

import "time"

type Fichaje struct {
    ID         uint       `gorm:"primaryKey" json:"id"`
    EmpleadoID uint       `json:"empleado_id"`
    Empleado   Empleado   `gorm:"foreignKey:EmpleadoID" json:"empleado,omitempty"`

    Entrada    time.Time  `json:"-"`
    Salida     *time.Time `json:"-"`

    // Campos auxiliares para formato legible
    EntradaStr string `gorm:"-" json:"entrada"`
    SalidaStr  string `gorm:"-" json:"salida,omitempty"`

	// Nuevos campos para coordenadas GPS
    Latitud  float64 `json:"latitud"`
    Longitud float64 `json:"longitud"`

    Ubicacion  string     `json:"ubicacion"`
}

// Después definimos un método que GORM no usa, pero nosotros llamaremos antes de devolver:
func (f *Fichaje) FormatearFechas() {
    f.EntradaStr = f.Entrada.Format("2006-01-02 15:04:05")
    if f.Salida != nil {
        f.SalidaStr = f.Salida.Format("2006-01-02 15:04:05")
    }
}

