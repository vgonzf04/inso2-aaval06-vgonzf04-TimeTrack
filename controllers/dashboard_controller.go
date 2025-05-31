package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// --------------------------------------------------------------------------------------------------------------------
// 1. HorasTrabajadasPorPeriodo
// GET /dashboard/horas-periodo
// Query params obligatorios: inicio=YYYY-MM-DD, fin=YYYY-MM-DD
// Query params opcionales: empleado_id=<num>, supervisor_id=<num>
// Respuesta: [ { "empleado_id":1, "nombre":"Juan Pérez", "total_horas":160.5 }, ... ]
func HorasTrabajadasPorPeriodo(c *gin.Context) {
	// 1. Leer y validar parámetros "inicio" y "fin"
	inicioStr := c.Query("inicio")
	finStr := c.Query("fin")
	if inicioStr == "" || finStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe especificar 'inicio' y 'fin' en formato YYYY-MM-DD"})
		return
	}
	inicio, err1 := time.Parse("2006-01-02", inicioStr)
	fin, err2 := time.Parse("2006-01-02", finStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Debe ser YYYY-MM-DD"})
		return
	}
	// Ajustar fin para incluir todo el día "fin"
	fin = fin.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// 2. Definir struct para resultados
	type Resultado struct {
		EmpleadoID uint    `json:"empleado_id"`
		Nombre     string  `json:"nombre"`
		TotalHoras float64 `json:"total_horas"`
	}
	var resultados []Resultado

	// 3. Construir consulta base
	db := config.DB.
		Table("fichajes AS f").
		Select("f.empleado_id, e.nombre, SUM(EXTRACT(EPOCH FROM (f.salida - f.entrada)) / 3600) AS total_horas").
		Joins("JOIN empleados AS e ON e.id = f.empleado_id").
		Where("f.entrada >= ? AND f.entrada <= ? AND f.salida IS NOT NULL", inicio, fin)

	// 4. Filtros opcionales
	if empID := c.Query("empleado_id"); empID != "" {
		id, err := strconv.Atoi(empID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empleado_id debe ser un número"})
			return
		}
		db = db.Where("f.empleado_id = ?", id)
	}
	if supID := c.Query("supervisor_id"); supID != "" {
		id, err := strconv.Atoi(supID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "supervisor_id debe ser un número"})
			return
		}
		// Unimos de nuevo con la tabla empleados (e2) para filtrar por supervisor_id
		db = db.Joins("JOIN empleados AS e2 ON e2.id = f.empleado_id").
			Where("e2.supervisor_id = ?", id)
	}

	// 5. Agrupar por empleado y ejecutar consulta
	if err := db.
		Group("f.empleado_id, e.nombre").
		Scan(&resultados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al calcular horas trabajadas"})
		return
	}

	// 6. Devolver JSON
	c.JSON(http.StatusOK, resultados)
}

// --------------------------------------------------------------------------------------------------------------------
// 2. PromedioHorasPorPeriodo
// GET /dashboard/promedio-horas
// Query params obligatorios: inicio=YYYY-MM-DD, fin=YYYY-MM-DD, periodo=diario|semanal|mensual
// Query params opcionales: empleado_id=<num>, supervisor_id=<num>
// Respuesta: [ { "empleado_id":1, "nombre":"Juan Pérez", "promedio_horas":5.35 }, ... ]
func PromedioHorasPorPeriodo(c *gin.Context) {
	inicioStr := c.Query("inicio")
	finStr := c.Query("fin")
	periodo := c.Query("periodo")
	if inicioStr == "" || finStr == "" || periodo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe indicar 'inicio', 'fin' y 'periodo' (diario|semanal|mensual)."})
		return
	}
	// Parsear fechas
	inicio, err1 := time.Parse("2006-01-02", inicioStr)
	fin, err2 := time.Parse("2006-01-02", finStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fechas inválido. Debe ser YYYY-MM-DD."})
		return
	}
	// Ajustar fin para incluir todo el día
	fin = fin.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// Calcular número de días completos en el rango
	duracionDias := fin.Sub(inicio).Hours()/24 + 1
	if duracionDias <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'fin' debe ser igual o posterior a 'inicio'."})
		return
	}

	// Determinar el divisor según el periodo
	var divisor float64
	switch periodo {
	case "diario":
		divisor = duracionDias
	case "semanal":
		divisor = duracionDias / 7.0
	case "mensual":
		yearIni, monIni, _ := inicio.Date()
		yearFin, monFin, _ := fin.Date()
		meses := float64((yearFin-yearIni)*12 + int(monFin) - int(monIni) + 1)
		if meses <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rango inválido para cálculo mensual."})
			return
		}
		divisor = meses
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Periodo inválido. Debe ser 'diario', 'semanal' o 'mensual'."})
		return
	}

	// Definir estructura temporal para resultados totales
	type Temp struct {
		EmpleadoID uint
		Nombre     string
		TotalHoras float64
	}
	var tmpResults []Temp

	// Construir consulta similar a HorasTrabajadasPorPeriodo
	db := config.DB.
		Table("fichajes AS f").
		Select("f.empleado_id, e.nombre, SUM(EXTRACT(EPOCH FROM (f.salida - f.entrada)) / 3600) AS total_horas").
		Joins("JOIN empleados AS e ON e.id = f.empleado_id").
		Where("f.entrada >= ? AND f.entrada <= ? AND f.salida IS NOT NULL", inicio, fin)

	// Filtros opcionales
	if empID := c.Query("empleado_id"); empID != "" {
		id, err := strconv.Atoi(empID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empleado_id debe ser numérico."})
			return
		}
		db = db.Where("f.empleado_id = ?", id)
	}
	if supID := c.Query("supervisor_id"); supID != "" {
		id, err := strconv.Atoi(supID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "supervisor_id debe ser numérico."})
			return
		}
		db = db.Joins("JOIN empleados AS e2 ON e2.id = f.empleado_id").
			Where("e2.supervisor_id = ?", id)
	}

	// Ejecutar consulta grupal
	if err := db.
		Group("f.empleado_id, e.nombre").
		Scan(&tmpResults).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al calcular total de horas"})
		return
	}

	// Construir resultado final con promedio
	type Resultado struct {
		EmpleadoID    uint    `json:"empleado_id"`
		Nombre        string  `json:"nombre"`
		PromedioHoras float64 `json:"promedio_horas"`
	}
	var resultados []Resultado
	for _, r := range tmpResults {
		promedio := r.TotalHoras / divisor
		resultados = append(resultados, Resultado{
			EmpleadoID:    r.EmpleadoID,
			Nombre:        r.Nombre,
			PromedioHoras: promedio,
		})
	}

	c.JSON(http.StatusOK, resultados)
}

// --------------------------------------------------------------------------------------------------------------------
// 3. FichajesPorDia
// GET /dashboard/fichajes-dia
// Query param: dia=YYYY-MM-DD
// Respuesta: { "dia":"YYYY-MM-DD", "fichajes_abiertos":X, "fichajes_cerrados":Y }
func FichajesPorDia(c *gin.Context) {
	diaStr := c.Query("dia")
	if diaStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe indicar 'dia' en formato YYYY-MM-DD"})
		return
	}
	dia, err := time.Parse("2006-01-02", diaStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de 'dia' inválido. Debe ser YYYY-MM-DD"})
		return
	}
	// Definir inicioDia (00:00:00) y finDia (23:59:59)
	inicioDia := dia
	finDia := dia.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// Contar fichajes abiertos (entrada en el día y salida IS NULL)
	var abiertos int64
	if err := config.DB.
		Model(&models.Fichaje{}).
		Where("entrada >= ? AND entrada <= ? AND salida IS NULL", inicioDia, finDia).
		Count(&abiertos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar fichajes abiertos"})
		return
	}

	// Contar fichajes cerrados (salida en el día)
	var cerrados int64
	if err := config.DB.
		Model(&models.Fichaje{}).
		Where("salida >= ? AND salida <= ?", inicioDia, finDia).
		Count(&cerrados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar fichajes cerrados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dia":               diaStr,
		"fichajes_abiertos": abiertos,
		"fichajes_cerrados": cerrados,
	})
}

// --------------------------------------------------------------------------------------------------------------------
// 4. VacacionesPorEstado
// GET /dashboard/vacaciones-por-estado
// Respuesta: [ { "estado":"pendiente", "cantidad":5 }, { "estado":"aprobada", "cantidad":12 }, ... ]
func VacacionesPorEstado(c *gin.Context) {
	type Resultado struct {
		Estado   string `json:"estado"`
		Cantidad int64  `json:"cantidad"`
	}
	var resultados []Resultado

	if err := config.DB.
		Model(&models.Vacacion{}).
		Select("estado, COUNT(*) AS cantidad").
		Group("estado").
		Scan(&resultados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar solicitudes de vacaciones por estado"})
		return
	}

	c.JSON(http.StatusOK, resultados)
}

// --------------------------------------------------------------------------------------------------------------------
// 5. VacacionesConsumidasPorRango
// GET /dashboard/vacaciones-consumidas-rango
// Query params obligatorios: desde=YYYY-MM-DD, hasta=YYYY-MM-DD
// Query param opcional: empleado_id=<num>
// Respuesta: [ { "empleado_id":1, "nombre":"Juan Pérez", "dias_consumidos":10 }, ... ]
func VacacionesConsumidasPorRango(c *gin.Context) {
	desdeStr := c.Query("desde")
	hastaStr := c.Query("hasta")
	if desdeStr == "" || hastaStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe indicar 'desde' y 'hasta' (YYYY-MM-DD)"})
		return
	}
	desde, err1 := time.Parse("2006-01-02", desdeStr)
	hasta, err2 := time.Parse("2006-01-02", hastaStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Debe ser YYYY-MM-DD"})
		return
	}
	if hasta.Before(desde) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'hasta' debe ser igual o posterior a 'desde'"})
		return
	}
	// Ajustar fin al final del día "hasta"
	fin := hasta.AddDate(0, 0, 1).Add(-time.Nanosecond)

	type Resultado struct {
		EmpleadoID     uint   `json:"empleado_id"`
		Nombre         string `json:"nombre"`
		DiasConsumidos int    `json:"dias_consumidos"`
	}
	var resultados []Resultado

	// Consulta para sumar días de vacaciones aprobadas con solapamiento en [desde, fin]
	db := config.DB.
		Table("vacaciones AS v").
		Select(`
			v.empleado_id,
			e.nombre,
			SUM(
				(
					LEAST(v.fin + INTERVAL '1 day', ?)::DATE
					- GREATEST(v.inicio, ?)::DATE
				)
			) AS dias_consumidos
		`, fin, desde).
		Joins("JOIN empleados AS e ON e.id = v.empleado_id").
		Where("v.estado = 'aprobada' AND v.fin >= ? AND v.inicio <= ?", desde, fin)

	// Filtrar opcionalmente por empleado_id
	if empID := c.Query("empleado_id"); empID != "" {
		id, err := strconv.Atoi(empID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "empleado_id debe ser un número"})
			return
		}
		db = db.Where("v.empleado_id = ?", id)
	}

	// Agrupar y ejecutar
	if err := db.
		Group("v.empleado_id, e.nombre").
		Scan(&resultados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al calcular días de vacaciones"})
		return
	}

	c.JSON(http.StatusOK, resultados)
}
