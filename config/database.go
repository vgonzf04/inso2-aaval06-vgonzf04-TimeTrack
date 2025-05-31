package config

import (
    "fmt"
    "log"
    "os"
    "github.com/cenkalti/backoff/v4"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "AppWebPruebaEmpleados/models"
)

var DB *gorm.DB

func ConectarBD() {
    host := os.Getenv("DB_HOST")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    port := os.Getenv("DB_PORT")

    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        host, user, password, dbname, port,
    )

    var db *gorm.DB
    operation := func() error {
        var err error
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        return err
    }

    err := backoff.Retry(operation, backoff.NewExponentialBackOff())
    if err != nil {
        log.Fatalf("❌ Error al conectar con la base de datos: %v", err)
    }

    DB = db
    fmt.Println("✅ Conexión a la base de datos establecida")

    err = DB.AutoMigrate(&models.Empleado{}, &models.Fichaje{}, &models.Vacacion{})
    if err != nil {
        log.Fatalf("❌ Error al migrar modelos: %v", err)
    }

    fmt.Println("✅ Migración completada")
}
