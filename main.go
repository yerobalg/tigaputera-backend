package main

import (
	"tigaputera-backend/src/database"
	"tigaputera-backend/src/routes"
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

	_, err := database.Init(logger)
	if err != nil {
		panic(err)
	}

	r := routes.Init(logger)
	r.Run()
}
