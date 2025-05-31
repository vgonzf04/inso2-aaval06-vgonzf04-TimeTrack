package controllers

import (
	"net/http"
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// ObtenerTodosEmpleados devuelve la lista completa de empleados
func ObtenerTodosEmpleados(c *gin.Context) {
	var empleados []models.Empleado
	result := config.DB.Find(&empleados)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, empleados)
}

// ObtenerEmpleadoPorID devuelve un empleado por su ID
func ObtenerEmpleadoPorID(c *gin.Context) {
	id := c.Param("id")
	var empleado models.Empleado

	if err := config.DB.First(&empleado, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
		return
	}

	c.JSON(http.StatusOK, empleado)
}

// CrearEmpleado añade un nuevo empleado
func CrearEmpleado(c *gin.Context) {
	var nuevo models.Empleado
	if err := c.ShouldBindJSON(&nuevo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	nuevo.FechaContratacion = time.Now().Format("2006-01-02")

	// Si no se especifica el rol o se envía vacío, asignar por defecto "Empleado"
	rol := strings.ToLower(strings.TrimSpace(nuevo.Rol))

	if rol == "" || rol == "empleado" {
		nuevo.Rol = "Empleado"
	} else if rol == "supervisor" {
		nuevo.Rol = "Supervisor"
		nuevo.SupervisorID = nil
	}else{
		// Si se envía un rol inválido, podrías rechazarlo también
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rol inválido. Solo se permite 'Empleado' o 'Supervisor'"})
		return
	}

	// ✅ Si el rol es "Empleado", validar que el supervisor exista y sea Supervisor
	if nuevo.Rol == "Empleado" && nuevo.SupervisorID != nil {
		var supervisor models.Empleado
		if err := config.DB.First(&supervisor, *nuevo.SupervisorID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El supervisor indicado no existe"})
			return
		}
		if strings.ToLower(supervisor.Rol) != "supervisor" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El empleado asignado como supervisor no tiene rol de 'Supervisor'"})
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
	id := c.Param("id")
	var empleado models.Empleado

	if err := config.DB.First(&empleado, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
		return
	}

	var datosActualizados models.Empleado
	if err := c.ShouldBindJSON(&datosActualizados); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Actualizar campos
	empleado.Nombre = datosActualizados.Nombre
	empleado.Email = datosActualizados.Email
	empleado.Cargo = datosActualizados.Cargo
	empleado.FechaContratacion = datosActualizados.FechaContratacion
	empleado.SupervisorID = datosActualizados.SupervisorID

	config.DB.Save(&empleado)

	c.JSON(http.StatusOK, empleado)
}

// EliminarEmpleado borra un empleado por su ID
func EliminarEmpleado(c *gin.Context) {
	id := c.Param("id")
	var empleado models.Empleado

	if err := config.DB.First(&empleado, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
		return
	}

	config.DB.Delete(&empleado)
	c.JSON(http.StatusOK, gin.H{"mensaje": "Empleado eliminado correctamente"})
}
