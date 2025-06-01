package controllers

import (
	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
	"net/http"
	"strconv"
	"strings"
	"time"
	"fmt"
	"github.com/gin-gonic/gin"
)

// ObtenerEmpleadoPorID devuelve un empleado por su ID
func ObtenerEmpleadoPorID(c *gin.Context) {
	// 1) Recuperar usuario_id y rol_usuario del contexto (JWTAuth los puso allí)
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

	// 2) Si el rol NO es "supervisor", denegar (403)
	if rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden consultar empleados por ID"})
		return
	}

	// 3) Obtener el ID del empleado solicitado de la URL
	paramID := c.Param("id")
	empIDUint64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	empleadoID := uint(empIDUint64)

	// 4) Intentar cargar de la BD al empleado cuyo id = empleadoID
	//    y cuyo supervisor_id = usuarioID
	var empleado models.Empleado
	result := config.DB.
		Where("id = ? AND supervisor_id = ?", empleadoID, usuarioID).
		First(&empleado)

	if result.Error != nil {
		// Si no existe o no pertenece al supervisor, devolvemos 404 Not Found
		// (o podrías decidir enviar 403 Forbidden, pero convención es 404 para “no existe o no tienes acceso”)
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado o no pertenece a este supervisor"})
		return
	}

	// 5) Finalmente, devolver el empleado encontrado
	c.JSON(http.StatusOK, empleado)
}

// ObtenerTodosEmpleados devuelve la lista completa de empleados
func ObtenerTodosEmpleados(c *gin.Context) {
	// 1) Extraer del contexto los valores agregados por JWTAuth():
	//    - "usuario_id" en claims["sub"]
	//    - "rol_usuario" en claims["rol"]
	idRaw, existsID := c.Get("usuario_id")
	rolRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		// No debería ocurrir si JWTAuth está funcionando correctamente
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	// Convertir los valores a tipos concretos
	usuarioID, okID := idRaw.(uint)
	rolUsuario, okRol := rolRaw.(string)
	if !okID || !okRol {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al leer contexto"})
		return
	}

	var empleados []models.Empleado

	switch rolUsuario {
	case "supervisor":
		// 2) Si es supervisor, listar solo los empleados que él supervisa
		if err := config.DB.
			Where("supervisor_id = ?", usuarioID).
			Find(&empleados).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar empleados"})
			return
		}
	case "empleado":
		// 3) Si es empleado, devolver únicamente su propio registro
		var self models.Empleado
		if err := config.DB.First(&self, usuarioID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar su perfil"})
			return
		}
		empleados = append(empleados, self)
	default:
		// 4) Si hubiera otro rol (o no estuviera presente), denegar
		c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para listar empleados"})
		return
	}

	// 5) Devolver la lista (vacía o con los registros filtrados)
	c.JSON(http.StatusOK, empleados)
}

// CrearEmpleado añade un nuevo empleado
func CrearEmpleado(c *gin.Context) {
	var nuevo models.Empleado
	if err := c.ShouldBindJSON(&nuevo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	nuevo.FechaContratacion = time.Now().Format("2006-01-02")

	fmt.Printf("✅ JSON recibido: %+v\n", nuevo)

	// Si no se especifica el rol o se envía vacío, asignar por defecto "Empleado"
	rol := strings.ToLower(strings.TrimSpace(nuevo.Rol))

	if rol == "" || rol == "empleado" {
		nuevo.Rol = "empleado"
	} else if rol == "supervisor" {
		nuevo.Rol = "supervisor"
		nuevo.SupervisorID = nil
	} else {
		// Si se envía un rol inválido, podrías rechazarlo también
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rol inválido. Solo se permite 'Empleado' o 'Supervisor'"})
		return
	}

	// ✅ Si el rol es "Empleado", validar que el supervisor exista y sea Supervisor
	if nuevo.Rol == "empleado" && nuevo.SupervisorID != nil {
		var supervisor models.Empleado
		if err := config.DB.First(&supervisor, *nuevo.SupervisorID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El supervisor indicado no existe"})
			return
		}
		if strings.ToLower(supervisor.Rol) != "supervisor" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El empleado asignado como supervisor no tiene rol de 'supervisor'"})
			return
		}

		result := config.DB.Create(&nuevo)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusCreated, nuevo)
	}
}

// ActualizarEmpleado actualiza un empleado existente
func ActualizarEmpleado(c *gin.Context) {
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

	// 2) Solo supervisores pueden actualizar empleados
	if rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden actualizar empleados"})
		return
	}

	// 3) Parsear el ID del empleado que se quiere actualizar
	paramID := c.Param("id")
	empIDUint64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	empleadoID := uint(empIDUint64)

	// 4) Cargar de BD al empleado, pero solo si su supervisor_id = usuarioID
	var empleado models.Empleado
	result := config.DB.
		Where("id = ? AND supervisor_id = ?", empleadoID, usuarioID).
		First(&empleado)
	if result.Error != nil {
		// No existe o no está asignado a este supervisor
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado o no pertenece a este supervisor"})
		return
	}

	// 5) Leer JSON entrante en struct auxiliar
	var datos struct {
		Nombre            string `json:"nombre"`
		Email             string `json:"email"`
		Cargo             string `json:"cargo"`
		FechaContratacion string `json:"fecha_contratacion"`
		SupervisorID      *uint  `json:"supervisor_id"`
		Rol               string `json:"rol"`
	}
	if err := c.ShouldBindJSON(&datos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// 6) Actualizar campos (solo si vienen no vacíos)
	if datos.Nombre != "" {
		empleado.Nombre = datos.Nombre
	}
	if datos.Email != "" {
		empleado.Email = datos.Email
	}
	if datos.Cargo != "" {
		empleado.Cargo = datos.Cargo
	}
	if datos.FechaContratacion != "" {
		empleado.FechaContratacion = datos.FechaContratacion
	}
	// Cambiar supervisor_id (puede ser null o nuevo ID)
	empleado.SupervisorID = datos.SupervisorID

	// 7) Actualizar el rol (solo supervisor puede cambiar rol)
	if datos.Rol != "" {
		rolNuevo := strings.ToLower(datos.Rol)
		if rolNuevo != "empleado" && rolNuevo != "supervisor" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rol inválido. Solo 'empleado' o 'supervisor'"})
			return
		}
		empleado.Rol = rolNuevo
	}

	// 8) Guardar cambios en la BD
	if err := config.DB.Save(&empleado).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar cambios"})
		return
	}

	// 9) Devolver el empleado actualizado
	c.JSON(http.StatusOK, empleado)
}

// EliminarEmpleado borra un empleado por su ID
func EliminarEmpleado(c *gin.Context) {
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

	// 2) Solo supervisores pueden eliminar empleados
	if rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado: solo supervisores pueden eliminar empleados"})
		return
	}

	// 3) Parsear el ID del empleado a eliminar
	paramID := c.Param("id")
	empIDUint64, err := strconv.ParseUint(paramID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	empleadoID := uint(empIDUint64)

	// 4) Intentar cargar al empleado con id = empleadoID y supervisor_id = usuarioID
	var empleado models.Empleado
	result := config.DB.
		Where("id = ? AND supervisor_id = ?", empleadoID, usuarioID).
		First(&empleado)
	if result.Error != nil {
		// No existe o no le pertenece a este supervisor
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado o no pertenece a este supervisor"})
		return
	}

	// 5) Borrar el empleado de la BD
	if err := config.DB.Delete(&empleado).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el empleado"})
		return
	}

	// 6) Responder con 204 No Content o con mensaje de confirmación
	c.JSON(http.StatusNoContent, nil)
	// Alternativamente, podrías usar:
	// c.JSON(http.StatusOK, gin.H{"message": "Empleado eliminado correctamente"})
}

// ObtenerUsuarioAutenticado devuelve los datos del usuario autenticado
func ObtenerUsuarioAutenticado(c *gin.Context) {
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

	// 2) Cargar el empleado de la BD por su ID
	var empleado models.Empleado
	if err := config.DB.First(&empleado, usuarioID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar su perfil"})
		return
	}

	// 3) Devolver los datos del empleado (sin contraseña ni datos sensibles)
	c.JSON(http.StatusOK, gin.H{
		"id":                 empleado.ID,
		"nombre":             empleado.Nombre,
		"email":              empleado.Email,
		"cargo":              empleado.Cargo,
		"fecha_contratacion": empleado.FechaContratacion,
		"supervisor_id":      empleado.SupervisorID,
		"rol":                rolUsuario, // Devolver el rol del usuario autenticado
	})
}
