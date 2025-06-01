package controllers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "strconv"

    "AppWebPruebaEmpleados/config"
    "AppWebPruebaEmpleados/models"
)

// CrearVacacion crea una solicitud de vacaciones.
// - Si quien la solicita es supervisor, el estado se establece en "aprobada".  
// - Si quien la solicita es empleado, queda en "pendiente".  
// Además:
// 1) Un supervisor puede crear para sí mismo o para cualquiera de sus subordinados.  
// 2) Ningún usuario (empleado o supervisor) puede pedir vacaciones que se solapen con un período ya aprobado del mismo empleado.  
// 3) En la respuesta, las fechas ("inicio" y "fin") se devuelven únicamente en formato "YYYY-MM-DD" (sin hora).
func CrearVacacion(c *gin.Context) {
    // 1. Leer JSON de entrada
    var input struct {
        EmpleadoID uint   `json:"empleado_id" binding:"required"`
        Inicio     string `json:"inicio" binding:"required"` // espera "YYYY-MM-DD"
        Fin        string `json:"fin" binding:"required"`    // espera "YYYY-MM-DD"
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido o faltan campos"})
        return
    }

    // 2. Parsear las fechas sin hora
    fechaInicio, err1 := time.Parse("2006-01-02", input.Inicio)
    fechaFin, err2 := time.Parse("2006-01-02", input.Fin)
    if err1 != nil || err2 != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Debe ser YYYY-MM-DD"})
        return
    }
    // Asegurar que fechaFin >= fechaInicio
    if fechaFin.Before(fechaInicio) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "`fin` debe ser igual o posterior a `inicio`"})
        return
    }

    // 3. Verificar que el empleado existe
    var emp models.Empleado
    if err := config.DB.First(&emp, input.EmpleadoID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar empleado"})
        }
        return
    }

    // 4. Extraer rol y usuario_id del contexto
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

    // 5. Si es supervisor, verificar permiso para crear vacación para el EmpleadoID dado
    if rolUsuario == "supervisor" {
        if input.EmpleadoID != usuarioID {
            // Si no es para sí mismo, debe ser un subordinado
            var subordinado models.Empleado
            err := config.DB.
                Where("id = ? AND supervisor_id = ?", input.EmpleadoID, usuarioID).
                First(&subordinado).Error
            if err != nil {
                if err == gorm.ErrRecordNotFound {
                    c.JSON(http.StatusForbidden, gin.H{"error": "No puedes crear vacación para un empleado que no te pertenece"})
                } else {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar empleado subordinado"})
                }
                return
            }
        }
    } else if rolUsuario != "empleado" {
        // Cualquier otro rol que no sea “empleado” ni “supervisor” no está autorizado
        c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para crear vacación"})
        return
    }

    // 6. Validar solapamiento con vacacione s Aprobadas del mismo empleado
    var existente models.Vacacion
    err := config.DB.
        Where("empleado_id = ? AND estado = 'aprobada' AND NOT (fin < ? OR inicio > ?)",
            input.EmpleadoID, fechaInicio, fechaFin).
        First(&existente).Error
    if err == nil {
        // Encontró una vacación aprobada que se solapa
        c.JSON(http.StatusBadRequest, gin.H{"error": "Ya existe una vacación aprobada que se solapa en esas fechas"})
        return
    }
    if err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al comprobar vacacione s existentes"})
        return
    }

    // 7. Determinar el estado inicial según el rol
    estadoInicial := "pendiente"
    if rolUsuario == "supervisor" {
        estadoInicial = "aprobada"
    }

    // 8. Crear la vacación en la BD
    v := models.Vacacion{
        EmpleadoID: input.EmpleadoID,
        Inicio:     fechaInicio,
        Fin:        fechaFin,
        Estado:     estadoInicial,
    }
    if err := config.DB.Create(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la solicitud de vacaciones"})
        return
    }

    // 9. Preload del empleado para la respuesta
    config.DB.Preload("Empleado").First(&v, v.ID)

    // 10. Enviar JSON con fecha solo "YYYY-MM-DD" (sin hora)
    c.JSON(http.StatusCreated, gin.H{
        "id":          v.ID,
        "empleado_id": v.EmpleadoID,
        "empleado":    v.Empleado,
        "inicio":      input.Inicio,
        "fin":         input.Fin,
        "estado":      v.Estado,
    })
}

// ListarVacaciones devuelve todas las solicitudes (puede filtrar por empleado_id, estado, rango fechas)
// ListarVacaciones devuelve las vacaciones según el rol del usuario autenticado:
// - Si es "supervisor": trae todas las vacacione s de sus empleados asignados (e.supervisor_id = usuarioID)
//   además de sus propias vacacione s (vacacion.empleado_id = usuarioID).
// - Si es "empleado": solo trae las vacacione s donde vacacion.empleado_id = usuarioID.
// Adicionalmente, acepta filtros opcionales: empleado_id, estado, desde, hasta.
func ListarVacaciones(c *gin.Context) {
    // 1) Obtener usuario_id y rol_usuario del contexto (provenientes de JWTAuth)
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

    var vacas []models.Vacacion
    db := config.DB.Preload("Empleado")

    switch rolUsuario {
    case "supervisor":
        // 2.a) Supervisor: vacacione s propias o de empleados cuyo supervisor_id = usuarioID
        // Hacemos JOIN con empleados e para filtrar e.supervisor_id = usuarioID,
        // o vacacion.empleado_id = usuarioID
        db = db.Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
            Where("e.supervisor_id = ? OR vacacions.empleado_id = ?", usuarioID, usuarioID)
    case "empleado":
        // 2.b) Empleado normal: solo sus propias vacacione s
        db = db.Where("empleado_id = ?", usuarioID)
    default:
        c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para listar vacaciones"})
        return
    }

    // 3) Aplicar filtros opcionales (solo si vinieron en query params)
    if empleadoIDStr := c.Query("empleado_id"); empleadoIDStr != "" {
        if empIDUint64, err := strconv.ParseUint(empleadoIDStr, 10, 32); err == nil {
            empID := uint(empIDUint64)
            db = db.Where("empleado_id = ?", empID)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'empleado_id' inválido"})
            return
        }
    }
    if estado := c.Query("estado"); estado != "" {
        db = db.Where("estado = ?", estado)
    }
    if desde := c.Query("desde"); desde != "" {
        if t, err := time.Parse("2006-01-02", desde); err == nil {
            db = db.Where("fin >= ?", t)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'desde' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }
    if hasta := c.Query("hasta"); hasta != "" {
        if t, err := time.Parse("2006-01-02", hasta); err == nil {
            db = db.Where("inicio <= ?", t)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'hasta' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }

    // 4) Ejecutar la consulta
    if err := db.Find(&vacas).Error; err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar solicitudes de vacaciones"})
        return
    }

    // 5) Responder con el slice (vacío si no hay registros)
    c.JSON(http.StatusOK, vacas)
}

// AprobarVacacion cambia el estado de la solicitud a "aprobada"
// AprobarVacacion permite que sólo un supervisor apruebe una solicitud
// de vacaciones si la solicitud está en estado "pendiente" y el empleado
// pertenece a ese supervisor. Un empleado normal recibe 403 Forbidden.
func AprobarVacacion(c *gin.Context) {
    // 1) Extraer usuario_id y rol_usuario del contexto
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

    // 2) Sólo supervisores pueden aprobar
    if rolUsuario != "supervisor" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden aprobar solicitudes"})
        return
    }

    // 3) Parsear el ID de la vacación
    paramID := c.Param("id")
    vacIDUint64, err := strconv.ParseUint(paramID, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
        return
    }
    vacacionID := uint(vacIDUint64)

    // 4) Cargar la vacación sólo si pertenece a un empleado asignado a este supervisor
    var v models.Vacacion
    result := config.DB.
        Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
        Where("vacacions.id = ? AND e.supervisor_id = ?", vacacionID, usuarioID).
        First(&v)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Solicitud no encontrada o no pertenece a sus empleados"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar la solicitud"})
        }
        return
    }

    // 5) Sólo si está en estado "pendiente" se permite aprobar
    if v.Estado != "pendiente" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Solo se pueden aprobar solicitudes en estado 'pendiente'"})
        return
    }

    // 6) Cambiar estado a "aprobada" y guardar
    v.Estado = "aprobada"
    if err := config.DB.Save(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al aprobar la solicitud"})
        return
    }

    // 7) Preload empleado para la respuesta
    config.DB.Preload("Empleado").First(&v, v.ID)

    c.JSON(http.StatusOK, v)
}

// RechazarVacacion permite que sólo un supervisor rechace una solicitud
// de vacaciones si la solicitud está en estado "pendiente" y el empleado
// pertenece a ese supervisor. Un empleado normal recibe 403 Forbidden.
func RechazarVacacion(c *gin.Context) {
    // 1) Extraer usuario_id y rol_usuario del contexto
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

    // 2) Sólo supervisores pueden rechazar
    if rolUsuario != "supervisor" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden rechazar solicitudes"})
        return
    }

    // 3) Parsear el ID de la vacación
    paramID := c.Param("id")
    vacIDUint64, err := strconv.ParseUint(paramID, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
        return
    }
    vacacionID := uint(vacIDUint64)

    // 4) Cargar la vacación sólo si pertenece a un empleado asignado a este supervisor
    var v models.Vacacion
    result := config.DB.
        Joins("JOIN empleados e ON e.id = vacacions.empleado_id").
        Where("vacacions.id = ? AND e.supervisor_id = ?", vacacionID, usuarioID).
        First(&v)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Solicitud no encontrada o no pertenece a sus empleados"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar la solicitud"})
        }
        return
    }

    // 5) Sólo si está en estado "pendiente" se permite rechazar
    if v.Estado != "pendiente" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Solo se pueden rechazar solicitudes en estado 'pendiente'"})
        return
    }

    // 6) Cambiar estado a "rechazada" y guardar
    v.Estado = "rechazada"
    if err := config.DB.Save(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al rechazar la solicitud"})
        return
    }

    // 7) Preload empleado para la respuesta
    config.DB.Preload("Empleado").First(&v, v.ID)

    c.JSON(http.StatusOK, v)
}
