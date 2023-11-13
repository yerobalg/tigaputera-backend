package database

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"tigaputera-backend/sdk/log"
)

type DB struct {
	*gorm.DB
}

func Init(dbLogger *log.Logger) (*DB, error) {
	db, err := initPostgres(dbLogger)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func initPostgres(dbLogger *log.Logger) (*gorm.DB, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		Logger: log.New(dbLogger),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	return db, nil
}