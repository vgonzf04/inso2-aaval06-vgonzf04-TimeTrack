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
// HorasTrabajadasPorPeriodo retorna las horas trabajadas en un rango, respetando el rol:
// - Supervisor: ve sus propias horas y las de sus empleados asignados.
// - Empleado: ve solo sus propias horas.
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

	// 3. Obtener usuario_id y rol_usuario del contexto (JWTAuth los puso)
	idRaw, existsID := c.Get("usuario_id")
	rolRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	usuarioID, okID := idRaw.(uint)
	rolUsuario, okRol := rolRaw.(string)
	if !okID || !okRol {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al leer contexto"})
		return
	}

	// 4. Empezar a construir la consulta base
	db := config.DB.
		Table("fichajes AS f").
		Select("f.empleado_id, e.nombre, SUM(EXTRACT(EPOCH FROM (f.salida - f.entrada)) / 3600) AS total_horas").
		Joins("JOIN empleados AS e ON e.id = f.empleado_id").
		Where("f.entrada >= ? AND f.entrada <= ? AND f.salida IS NOT NULL", inicio, fin)

	// 5. Aplicar restricciones según el rol
	switch rolUsuario {

	case "supervisor":
		// a) Si viene "empleado_id" en query, verificar que pertenece al supervisor o sea él mismo
		if empIDStr := c.Query("empleado_id"); empIDStr != "" {
			empIDUint64, err := strconv.ParseUint(empIDStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "empleado_id debe ser un número"})
				return
			}
			empID := uint(empIDUint64)

			// Verificar en BD que este empleado tiene supervisor_id = usuarioID o empID == usuarioID
			var tempEmp models.Empleado
			err = config.DB.
				Where("id = ? AND (supervisor_id = ? OR id = ?)", empID, usuarioID, usuarioID).
				First(&tempEmp).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusForbidden, gin.H{"error": "No puede ver horas de este empleado"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar empleado"})
				}
				return
			}
			db = db.Where("f.empleado_id = ?", empID)

		} else {
			// b) Si no viene "empleado_id", traer horas propias y de empleados asignados
			db = db.Where("e.supervisor_id = ? OR f.empleado_id = ?", usuarioID, usuarioID)
		}

	case "empleado":
		// Solo sus propias horas, ignorar filtro "empleado_id"
		db = db.Where("f.empleado_id = ?", usuarioID)

	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para ver horas trabajadas"})
		return
	}

	// 6. Ejecutar la consulta con agrupamiento
	if err := db.
		Group("f.empleado_id, e.nombre").
		Scan(&resultados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al calcular horas trabajadas"})
		return
	}

	// 7. Devolver resultados
	c.JSON(http.StatusOK, resultados)
}

// --------------------------------------------------------------------------------------------------------------------
// FichajesPorDia restringido a supervisores: cuenta fichajes de empleados asignados y del propio supervisor.
func FichajesPorDia(c *gin.Context) {
	// 1) Verificar rol y usuario_id del token
	idRaw, existsID := c.Get("usuario_id")
	rolRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	usuarioID, okID := idRaw.(uint)
	rolUsuario, okRol := rolRaw.(string)
	if !okID || !okRol {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al leer contexto"})
		return
	}
	if rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden ver éste reporte"})
		return
	}

	// 2) Leer y validar parámetro "dia"
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
	inicioDia := dia
	finDia := dia.AddDate(0, 0, 1).Add(-time.Nanosecond)

	// 3) Contar fichajes abiertos (entrada en el día, salida IS NULL) para sup y su equipo
	var abiertos int64
	err = config.DB.
		Model(&models.Fichaje{}).
		Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
		Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?) AND entrada >= ? AND entrada <= ? AND salida IS NULL",
			usuarioID, usuarioID, inicioDia, finDia).
		Count(&abiertos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar fichajes abiertos"})
		return
	}

	// 4) Contar fichajes cerrados (salida en el día) para sup y su equipo
	var cerrados int64
	err = config.DB.
		Model(&models.Fichaje{}).
		Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
		Where("(e.supervisor_id = ? OR fichajes.empleado_id = ?) AND salida >= ? AND salida <= ?",
			usuarioID, usuarioID, inicioDia, finDia).
		Count(&cerrados).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar fichajes cerrados"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dia":               diaStr,
		"fichajes_abiertos": abiertos,
		"fichajes_cerrados": cerrados,
	})
}

// VacacionesPorEstado restringido a supervisores: agrupa por estado solo vacacione s de empleados asignados y propias.
func VacacionesPorEstado(c *gin.Context) {
	// 1) Verificar rol y usuario_id del token
	idRaw, existsID := c.Get("usuario_id")
	rolRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	usuarioID, okID := idRaw.(uint)
	rolUsuario, okRol := rolRaw.(string)
	if !okID || !okRol {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al leer contexto"})
		return
	}
	if rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden ver éste reporte"})
		return
	}

	// 2) Definir struct para resultados
	type Resultado struct {
		Estado   string `json:"estado"`
		Cantidad int64  `json:"cantidad"`
	}
	var resultados []Resultado

	// 3) Consulta agrupada por estado de vacacione s de empleados del supervisor (y propias)
	if err := config.DB.
		Model(&models.Vacacion{}).
		Select("v.estado, COUNT(*) AS cantidad").
		Joins("JOIN empleados e ON e.id = v.empleado_id").
		Where("e.supervisor_id = ? OR v.empleado_id = ?", usuarioID, usuarioID).
		Group("v.estado").
		Scan(&resultados).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar solicitudes de vacaciones por estado"})
		return
	}

	c.JSON(http.StatusOK, resultados)
}
