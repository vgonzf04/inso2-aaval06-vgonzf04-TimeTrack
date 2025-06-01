package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"AppWebPruebaEmpleados/config"
	"AppWebPruebaEmpleados/models"
)

// CrearFichaje crea un nuevo registro de fichaje (marca la entrada)
// JSON esperado en el body: { "empleado_id": <int>, "latitud": <float>, "longitud": <float> }
func CrearFichaje(c *gin.Context) {
	idRaw, existsID := c.Get("usuario_id")
	rolRaw, existsRol := c.Get("rol_usuario")
	if !existsID || !existsRol {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	usuarioID, okID := idRaw.(uint)
	rolUsuario, okRol := rolRaw.(string)

	// Validar que el usuario tenga rol de empleado o supervisor
	if rolUsuario != "empleado" && rolUsuario != "supervisor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para crear fichajes"})
		return
	}
	var emp models.Empleado

	if err := config.DB.First(&emp, usuarioID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Empleado no encontrado"})
			return
		}
		log.Printf("Error al buscar empleado: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar empleado"})
		return
	}

	if !okID || !okRol {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno al leer contexto"})
		return
	}

	// 1. Struct para bindear el JSON entrante:
	var input struct {
		Latitud  float64 `json:"latitud"`
		Longitud float64 `json:"longitud"`
	}

	// 2. Intentar bindear JSON; si falla, devolvemos el error exacto
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Error al hacer ShouldBindJSON en CrearFichaje: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Validar que no exista ya un fichaje abierto para este empleado
	var abierto models.Fichaje
	err := config.DB.
		Where("empleado_id = ? AND salida IS NULL", emp.ID).
		First(&abierto).Error

	if err == nil {
		// Ya existe un fichaje abierto (no se cerró)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ya tienes un fichaje abierto. Cierra primero el anterior"})
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		// Algo falló al consultar la base de datos
		log.Printf("Error al comprobar fichajes abiertos: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al comprobar fichajes abiertos"})
		return
	}

	// 4. Obtener dirección mediante reverse-geocoding
	direccion := obtenerDireccionGoogle(input.Latitud, input.Longitud)

	// Cargar hora actual
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("No se pudo cargar zona Europe/Madrid, usando UTC: %v\n", err)
		loc = time.UTC
	}
	ahora := time.Now().In(loc)

	// 5. Crear el registro de fichaje con la hora de entrada actual
	f := models.Fichaje{
		EmpleadoID: emp.ID,
		Entrada:    ahora,
		Latitud:    input.Latitud,
		Longitud:   input.Longitud,
		Ubicacion:  direccion,
		Salida:     nil,
	}

	if err := config.DB.Create(&f).Error; err != nil {
		log.Printf("Error al crear fichaje en BD: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el fichaje"})
		return
	}

	// 6. *Recargar* el registro recién creado para traer la relación Empleado
	//    (con Preload cargamos todos los datos del empleado asociado).
	if err := config.DB.
		Preload("Empleado").
		First(&f, f.ID).Error; err != nil {
		log.Printf("Error al hacer Preload de Empleado: %v\n", err)
		// Aunque falle el Preload, podemos devolver el fichaje sin datos de empleado
	}

	// 7. Formatear las fechas (rellena f.EntradaStr y f.SalidaStr)
	f.FormatearFechas()

	// 8. Devolver el objeto completo en JSON (incluye datos de empleado y "entrada" formateada)
	c.JSON(http.StatusCreated, f)
}

// CerrarFichaje actualiza la hora de salida de un fichaje abierto
// No necesita body, se toma la hora actual como "salida".
func CerrarFichaje(c *gin.Context) {
	// 1. Obtener el parámetro :id de la URL
	id := c.Param("id")

	// 2. Buscar el fichaje por ID, incluyendo la relación Empleado
	var f models.Fichaje
	if err := config.DB.Preload("Empleado").First(&f, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fichaje no encontrado"})
		} else {
			log.Printf("Error al buscar fichaje (ID=%s): %v\n", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar el fichaje"})
		}
		return
	}

	// 3. Verificar si ya tiene salida asignada
	if f.Salida != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Este fichaje ya fue cerrado"})
		return
	}

	// 4. Cargar la zona horaria Europe/Madrid
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		log.Printf("No se pudo cargar zona Europe/Madrid, usando UTC: %v\n", err)
		loc = time.UTC
	}
	salida := time.Now().In(loc)

	// 5. Actualizar SOLO el campo Salida en la base de datos
	//    Usamos Updates(map[string]interface{}) para no sobreescribir otros campos
	if err := config.DB.Model(&f).Updates(map[string]interface{}{"salida": &salida}).Error; err != nil {
		log.Printf("Error al guardar la salida del fichaje (ID=%s): %v\n", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo cerrar el fichaje"})
		return
	}

	// 6. Recargar el registro para asegurarnos de tener la relación Empleado actualizada
	//    y el campo Salida en la estructura Go
	if err := config.DB.Preload("Empleado").First(&f, id).Error; err != nil {
		log.Printf("Error al recargar fichaje tras cerrar (ID=%s): %v\n", id, err)
		// Aunque falle el Preload, devolvemos el fichaje con la salida ya asignada
	}

	// 7. Formatear fechas (Entrada y Salida) antes de devolver
	f.FormatearFechas()

	// 8. Retornar el fichaje completo en JSON
	c.JSON(http.StatusOK, f)
}

func ObtenerFichajeActual(c *gin.Context) {
	// 1. Extraer usuario_id y rol_usuario del contexto (JWTAuth los puso)
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
	// 2. Buscar el fichaje abierto del usuario
	var fichaje models.Fichaje
	db := config.DB.Preload("Empleado")
	switch rolUsuario {
	case "supervisor":
		// 2a) Supervisor: busca fichajes abiertos de sus empleados y el suyo propio
		err := db.
			Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
			Where("e.supervisor_id = ? OR fichajes.empleado_id = ? AND fichajes.salida IS NULL", usuarioID, usuarioID).
			First(&fichaje).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "No hay fichaje abierto para este usuario"})
			} else {
				log.Printf("Error al buscar fichaje actual: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichaje actual"})
			}
			return
		}
	case "empleado":
		// 2b) Empleado normal: busca su propio fichaje abierto
		err := db.
			Where("empleado_id = ? AND salida IS NULL", usuarioID).
			First(&fichaje).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "No hay fichaje abierto para este usuario"})
			} else {
				log.Printf("Error al buscar fichaje actual: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichaje actual"})
			}
			return
		}
	default:
		// 2c) Cualquier otro rol no autorizado
		c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para consultar fichaje actual"})
		return
	}
}

// ListarFichajes obtiene el historial de fichajes
// Permite query params opcionales: empleado_id, desde (YYYY-MM-DD), hasta (YYYY-MM-DD)
func ListarFichajes(c *gin.Context) {
	// 1) Extraer usuario_id y rol_usuario del contexto (JWTAuth los puso)
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

	var fichajes []models.Fichaje
	db := config.DB.Preload("Empleado")

	switch rolUsuario {
	case "supervisor":
		// 2a) Supervisor: trae los fichajes de sus empleados y los suyos propios
		// Hacemos JOIN con la tabla empleados para filtrar por supervisor_id
		err := db.
			Joins("JOIN empleados e ON e.id = fichajes.empleado_id").
			Where("e.supervisor_id = ? OR fichajes.empleado_id = ?", usuarioID, usuarioID).
			Find(&fichajes).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichajes"})
			return
		}

	case "empleado":
		// 2b) Empleado normal: solo sus fichajes
		err := db.
			Where("empleado_id = ?", usuarioID).
			Find(&fichajes).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichajes"})
			return
		}

	default:
		// 2c) Cualquier otro rol (por si tuvieras más) no autorizado
		c.JSON(http.StatusForbidden, gin.H{"error": "Rol no autorizado para listar fichajes"})
		return
	}

	// 3) Devolver el slice de fichajes (vacío o con registros)
	c.JSON(http.StatusOK, fichajes)
}

// obtenerDireccionGoogle hace una petición al API de Google Maps Reverse Geocoding
// y devuelve la dirección textual (formatted_address) del primer resultado.
// Si algo falla, devuelve "".
func obtenerDireccionGoogle(lat, lng float64) string {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		return ""
	}
	url := fmt.Sprintf(
		"https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=%s",
		lat, lng, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error al llamar a Google Maps: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error al leer body de Google Maps: %v\n", err)
		return ""
	}

	var resultado struct {
		Results []struct {
			FormattedAddress string   `json:"formatted_address"`
			Types            []string `json:"types"`
		} `json:"results"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &resultado); err != nil {
		log.Printf("Error al parsear JSON de Google Maps: %v\n", err)
		return ""
	}
	if resultado.Status != "OK" || len(resultado.Results) == 0 {
		return ""
	}

	// Buscar el primer Result que tenga "street_address" en Types
	for _, r := range resultado.Results {
		for _, t := range r.Types {
			if t == "street_address" || t == "route" {
				return r.FormattedAddress
			}
		}
	}

	// Si no encontramos street_address, devolvemos el FormattedAddress del primer resultado
	return resultado.Results[0].FormattedAddress
}
