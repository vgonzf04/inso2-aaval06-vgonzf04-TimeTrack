package controllers

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"
    "log"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "AppWebPruebaEmpleados/config"
    "AppWebPruebaEmpleados/models"
)

// CrearFichaje crea un nuevo registro de fichaje (marca la entrada)
// JSON esperado en el body: { "empleado_id": <int>, "latitud": <float>, "longitud": <float> }
func CrearFichaje(c *gin.Context) {
    // 1. Struct para bindear el JSON entrante:
    var input struct {
        EmpleadoID uint    `json:"empleado_id"`
        Latitud    float64 `json:"latitud"`
        Longitud   float64 `json:"longitud"`
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
        Where("empleado_id = ? AND salida IS NULL", input.EmpleadoID).
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
        EmpleadoID: input.EmpleadoID,
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

// ListarFichajes obtiene el historial de fichajes
// Permite query params opcionales: empleado_id, desde (YYYY-MM-DD), hasta (YYYY-MM-DD)
func ListarFichajes(c *gin.Context) {
    var fichajes []models.Fichaje

    // 1. Construir la consulta base, haciendo Preload de Empleado
    query := config.DB.Preload("Empleado")

    // 2. (Opcional) si quieres filtrar por empleado_id o por rango de fechas, puedes descomentar:
    // if empID := c.Query("empleado_id"); empID != "" {
    //     id, err := strconv.Atoi(empID)
    //     if err == nil {
    //         query = query.Where("empleado_id = ?", id)
    //     }
    // }
    // if desde := c.Query("desde"); desde != "" {
    //     // parseamos "YYYY-MM-DD" a time.Time en Madrid
    //     loc, _ := time.LoadLocation("Europe/Madrid")
    //     tDesde, err := time.ParseInLocation("2006-01-02", desde, loc)
    //     if err == nil {
    //         query = query.Where("entrada >= ?", tDesde)
    //     }
    // }
    // if hasta := c.Query("hasta"); hasta != "" {
    //     loc, _ := time.LoadLocation("Europe/Madrid")
    //     tHasta, err := time.ParseInLocation("2006-01-02", hasta, loc)
    //     if err == nil {
    //         query = query.Where("entrada <= ?", tHasta)
    //     }
    // }

    // 3. Ejecutar la consulta
    if err := query.Find(&fichajes).Error; err != nil {
        log.Printf("Error al consultar fichajes: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al consultar fichajes"})
        return
    }

    // 4. Formatear fechas para cada fichaje antes de devolverlo
    for i := range fichajes {
        fichajes[i].FormatearFechas()
    }

    // 5. Devolver el array con todos los fichajes, ya formateados
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
