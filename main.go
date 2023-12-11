package main

import (
	"github.com/joho/godotenv"
	"tigaputera-backend/sdk/jwt"
	"tigaputera-backend/sdk/log"
	"tigaputera-backend/sdk/password"
	"tigaputera-backend/sdk/storage"
	"tigaputera-backend/sdk/validator"
	"tigaputera-backend/src/controller"
	"tigaputera-backend/src/database"

	"os"
)

// @title Tigaputera Backend API
// @description API about financial management for construction company
// @version 1.0

// @contact.name 	Yerobal Gustaf Sekeon
// @contact.email 	yerobalg@gmail.com

// @securitydefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @value Bearer {token}

func main() {
	loadEnv()
	initialize()
}

func loadEnv() {
	if os.Getenv("ENV") == "production" || os.Getenv("ENV") == "staging" {
		return
	}

	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func initialize() {
	logger := log.Init()

	validator := validator.Init()

	password := password.Init()

	jwt := jwt.Init()

	gcpServiceAccount := storage.GCPServiceAccount{
		Type:                    os.Getenv("SA_TYPE"),
		ProjectID:               os.Getenv("SA_PROJECT_ID"),
		PrivateKeyID:            os.Getenv("SA_PRIVATE_KEY_ID"),
		PrivateKey:              os.Getenv("SA_PRIVATE_KEY"),
		ClientEmail:             os.Getenv("SA_CLIENT_EMAIL"),
		ClientID:                os.Getenv("SA_CLIENT_ID"),
		AuthURI:                 os.Getenv("SA_AUTH_URI"),
		TokenURI:                os.Getenv("SA_TOKEN_URI"),
		AuthProviderX509CertURL: os.Getenv("SA_AUTH_PROVIDER_X509_CERT_URL"),
		ClientX509CertURL:       os.Getenv("SA_CLIENT_X509_CERT_URL"),
		UniverseDomain:          os.Getenv("SA_UNIVERSE_DOMAIN"),
	}

	storage := storage.Init(gcpServiceAccount, os.Getenv("STORAGE_BUCKET_NAME"))

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

	r := controller.Init(logger, db, jwt, password, validator, storage)
	r.Run()
}
