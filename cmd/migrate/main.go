package main

import (
	"fmt"
	"go-empty/pkg/config"
	"go-empty/pkg/db"
	"go-empty/pkg/models"
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const confPath = "./config.json"

func main() {

	cfg, err := config.New(confPath)
	if err != nil {
		log.Fatal(err)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	l := db.Logger{
		Logger: newLogger,
	}
	d, err := db.GormOpen(&cfg.DB, &l)
	if err != nil {
		log.Fatal(err)
	}

	migration(d)

}

func migration(db *gorm.DB) {

	fmt.Println("Run DB Migration --> Start")

	db.AutoMigrate(&models.Edge{})

	db.AutoMigrate(&models.EdgeStreaming{})

	db.AutoMigrate(&models.Device{})
	db.AutoMigrate(&models.EdgeReserve{})

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Platform{})
	db.AutoMigrate(&models.AppGenre{})
	db.AutoMigrate(&models.App{})
	db.AutoMigrate(&models.EdgeApp{})

	fmt.Println("Run DB Migration --> Done")

}
