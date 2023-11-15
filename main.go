package main

import (
	"tigaputera-backend/src/database"
	"tigaputera-backend/src/controller"
	"tigaputera-backend/sdk/log"

	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	initialize()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
}

func initialize() {
	logger := log.Init()

	db, err := database.Init(logger)
	if err != nil {
		panic(err)
	}

	if err := db.Migrate(); err != nil {
		panic(err)
	}

	if err := db.SeedSuperAdmin(); err != nil {
		panic(err)
	}

	r := controller.Init(logger, db)
	r.Run()
}
