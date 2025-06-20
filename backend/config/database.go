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

func ConnectDB() {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Madrid",
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
		log.Fatalf("❌ Failed to connect to the database: %v", err)
	}

	DB = db
	fmt.Println("✅ Database connection established")

	err = DB.AutoMigrate(&models.Employee{}, &models.TimeEntry{}, &models.Vacation{})
	if err != nil {
		log.Fatalf("❌ Failed to migrate models: %v", err)
	}

	fmt.Println("✅ Migration completed")
}
