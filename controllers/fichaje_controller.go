package controllers

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
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
    var input struct {
        EmpleadoID uint    `json:"empleado_id"`
        Latitud    float64 `json:"latitud"`
        Longitud   float64 `json:"longitud"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido o faltan campos"})
        return
    }

    // 1. Validar que no exista un fichaje abierto para este empleado
    var abierto models.Fichaje
    err := config.DB.
        Where("empleado_id = ? AND salida IS NULL", input.EmpleadoID).
        First(&abierto).Error

    if err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Ya tienes un fichaje abierto. Cierra primero el anterior"})
        return
    }
    if err != nil && err != gorm.ErrRecordNotFound {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al comprobar fichajes abiertos"})
        return
    }

    // 2. Hacer reverse-geocoding con Google Maps API para obtener dirección
    direccion := obtenerDireccionGoogle(input.Latitud, input.Longitud)

    // 3. Crear el registro de fichaje
    f := models.Fichaje{
        EmpleadoID: input.EmpleadoID,
        Entrada:    time.Now(),
        Latitud:    input.Latitud,
        Longitud:   input.Longitud,
        Ubicacion:  direccion, // cadena vacía si falla geocoding
        Salida:     nil,
    }

    if err := config.DB.Create(&f).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear el fichaje"})
        return
    }

    c.JSON(http.StatusCreated, f)
}

// CerrarFichaje actualiza la hora de salida de un fichaje abierto
// No necesita body, se toma la hora actual como "salida".
func CerrarFichaje(c *gin.Context) {
    id := c.Param("id")
    var f models.Fichaje

    // Buscar el fichaje por ID
    if err := config.DB.First(&f, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Fichaje no encontrado"})
        return
    }

    // Si ya tenía salida, no hacer nada (o reportar error)
    if f.Salida != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Este fichaje ya tiene hora de salida"})
        return
    }

    // Asignar hora de salida = ahora
    now := time.Now()
    f.Salida = &now

    if err := config.DB.Save(&f).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar la salida"})
        return
    }

    c.JSON(http.StatusOK, f)
}

// ListarFichajes obtiene el historial de fichajes
// Permite query params opcionales: empleado_id, desde (YYYY-MM-DD), hasta (YYYY-MM-DD)
func ListarFichajes(c *gin.Context) {
    var fichajes []models.Fichaje

    // Incluir datos del Empleado (opcional)
    query := config.DB.Preload("Empleado")

    // Filtrar por empleado_id si se proporciona
    if empleadoID := c.Query("empleado_id"); empleadoID != "" {
        query = query.Where("empleado_id = ?", empleadoID)
    }

    // Filtrar por rango de fecha (solo fecha de entrada)
    if desde := c.Query("desde"); desde != "" {
        if t, err := time.Parse("2006-01-02", desde); err == nil {
            query = query.Where("entrada >= ?", t)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'desde' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }
    if hasta := c.Query("hasta"); hasta != "" {
        if t, err := time.Parse("2006-01-02", hasta); err == nil {
            endOfDay := t.AddDate(0, 0, 1).Add(-time.Nanosecond)
            query = query.Where("entrada <= ?", endOfDay)
        } else {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetro 'hasta' inválido. Formato: YYYY-MM-DD"})
            return
        }
    }

    // Ejecutar la consulta
    if err := query.Find(&fichajes).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichajes"})
        return
    }

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
    // URL de Google: latlng=<lat>,<lng>&key=<API_KEY>
    url := fmt.Sprintf(
        "https://maps.googleapis.com/maps/api/geocode/json?latlng=%f,%f&key=%s",
        lat, lng, apiKey,
    )

    resp, err := http.Get(url)
    if err != nil {
        return ""
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return ""
    }

    var resultado struct {
        Results []struct {
            FormattedAddress string `json:"formatted_address"`
        } `json:"results"`
        Status string `json:"status"`
    }
    if err := json.Unmarshal(body, &resultado); err != nil {
        return ""
    }
    if resultado.Status != "OK" || len(resultado.Results) == 0 {
        return ""
    }
    return resultado.Results[0].FormattedAddress
}
