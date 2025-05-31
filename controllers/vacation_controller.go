package controllers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "AppWebPruebaEmpleados/config"
    "AppWebPruebaEmpleados/models"
)

// CrearVacacion permite a un empleado solicitar vacaciones.
// Recibe JSON: { "empleado_id": <int>, "inicio": "YYYY-MM-DD", "fin": "YYYY-MM-DD" }
// El estado inicial será "pendiente".
func CrearVacacion(c *gin.Context) {
    // 1. Cargar JSON de entrada
    var input struct {
        EmpleadoID uint   `json:"empleado_id"`
        Inicio     string `json:"inicio"`
        Fin        string `json:"fin"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido o faltan campos"})
        return
    }

    // 2. Parsear fechas "YYYY-MM-DD"
    fechaInicio, err1 := time.Parse("2006-01-02", input.Inicio)
    fechaFin, err2 := time.Parse("2006-01-02", input.Fin)
    if err1 != nil || err2 != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha inválido. Debe ser YYYY-MM-DD"})
        return
    }
    // 3. Validación: inicio <= fin
    if fechaFin.Before(fechaInicio) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "`fin` debe ser igual o posterior a `inicio`"})
        return
    }

    // 4. Verificar que el empleado existe
    var emp models.Empleado
    if err := config.DB.First(&emp, input.EmpleadoID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
        return
    }

    // 5. Validación extra: evitar solapamiento con otras solicitudes del mismo empleado
    //    Buscar si existe alguna vacacion con rango que se cruce y no esté RECHAZADA
    var existente models.Vacacion
    err := config.DB.
        Where("empleado_id = ? AND estado <> 'rechazada' AND NOT (fin < ? OR inicio > ?)",
            input.EmpleadoID, fechaInicio, fechaFin).
        First(&existente).Error

    if err == nil {
        // Encontró una solicitud activa (pendiente o aprobada) que se solapa
        c.JSON(http.StatusBadRequest, gin.H{"error": "Ya existe una solicitud de vacaciones que se solapa en esas fechas"})
        return
    }
    if err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al comprobar solicitudes existentes"})
        return
    }

    // 6. Crear la nueva solicitud de vacaciones con estado "pendiente"
    v := models.Vacacion{
        EmpleadoID: input.EmpleadoID,
        Inicio:     fechaInicio,
        Fin:        fechaFin,
        Estado:     "pendiente",
    }
    if err := config.DB.Create(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la solicitud de vacaciones"})
        return
    }

    // 7. (Opcional) Preload del empleado para devolver en la respuesta
    config.DB.Preload("Empleado").First(&v, v.ID)

    c.JSON(http.StatusCreated, v)
}

// ListarVacaciones devuelve todas las solicitudes (puede filtrar por empleado_id, estado, rango fechas)
func ListarVacaciones(c *gin.Context) {
    var vacas []models.Vacacion
    // 1. Iniciar query con Preload("Empleado") para incluir datos del empleado
    query := config.DB.Preload("Empleado")

    // 2. Filtros opcionales
    if empleadoID := c.Query("empleado_id"); empleadoID != "" {
        query = query.Where("empleado_id = ?", empleadoID)
    }
    if estado := c.Query("estado"); estado != "" {
        query = query.Where("estado = ?", estado)
    }
    if desde := c.Query("desde"); desde != "" {
        if t, err := time.Parse("2006-01-02", desde); err == nil {
            query = query.Where("fin >= ?", t)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'desde' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }
    if hasta := c.Query("hasta"); hasta != "" {
        if t, err := time.Parse("2006-01-02", hasta); err == nil {
            query = query.Where("inicio <= ?", t)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'hasta' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }

    // 3. Ejecutar consulta
    if err := query.Find(&vacas).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar solicitudes de vacaciones"})
        return
    }

    c.JSON(http.StatusOK, vacas)
}

// AprobarVacacion cambia el estado de la solicitud a "aprobada"
func AprobarVacacion(c *gin.Context) {
    id := c.Param("id")
    var v models.Vacacion

    // 1. Buscar la solicitud por ID
    if err := config.DB.First(&v, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Solicitud de vacaciones no encontrada"})
        return
    }

    // 2. Solo si está en estado "pendiente" se permite cambiar a "aprobada"
    if v.Estado != "pendiente" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Solo se pueden aprobar solicitudes en estado 'pendiente'"})
        return
    }

    // 3. Cambiar estado y guardar
    v.Estado = "aprobada"
    if err := config.DB.Save(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al aprobar la solicitud"})
        return
    }

    // 4. (Opcional) Preload empleado
    config.DB.Preload("Empleado").First(&v, v.ID)

    c.JSON(http.StatusOK, v)
}

// RechazarVacacion cambia el estado de la solicitud a "rechazada"
func RechazarVacacion(c *gin.Context) {
    id := c.Param("id")
    var v models.Vacacion

    // 1. Buscar la solicitud por ID
    if err := config.DB.First(&v, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Solicitud de vacaciones no encontrada"})
        return
    }

    // 2. Solo si está en estado "pendiente" se permite cambiar a "rechazada"
    if v.Estado != "pendiente" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Solo se pueden rechazar solicitudes en estado 'pendiente'"})
        return
    }

    // 3. Cambiar estado y guardar
    v.Estado = "rechazada"
    if err := config.DB.Save(&v).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al rechazar la solicitud"})
        return
    }

    // 4. (Opcional) Preload empleado
    config.DB.Preload("Empleado").First(&v, v.ID)

    c.JSON(http.StatusOK, v)
}
