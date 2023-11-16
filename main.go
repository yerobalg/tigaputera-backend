package main

import (
	"github.com/joho/godotenv"
	"tigaputera-backend/sdk/jwt"
	"tigaputera-backend/sdk/log"
	"tigaputera-backend/sdk/password"
	"tigaputera-backend/sdk/validator"
	"tigaputera-backend/src/controller"
	"tigaputera-backend/src/database"
)

// @title Tigaputera Backend API
// @description API about financial management for construction company
// @version 1.0

// @contact.name 	Yerobal Gustaf Sekeon
// @contact.email 	yerobalg@gmail.com

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

	validator := validator.Init()

	password := password.Init()

	jwt := jwt.Init()

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

	r := controller.Init(logger, db, jwt, password, validator)
	r.Run()
}
