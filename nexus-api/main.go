package main

import (
	"context"
	"fmt"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/logging"
	"nexus-api/password"
	"nexus-api/service"

	"os"
)

var (
	serviceCtx = context.Background()
)

func main() {
	// #TODO make into a unit test
	//generate a hash for password123
	hash, err := password.HashPassword("password123")
	if err != nil {
		fmt.Println("Error generating hash:", err)
		return
	}
	fmt.Printf("Hash for password123: %s\n", hash)

	// setup logger
	logLevel := os.Getenv("LOG_LEVEL")
	serviceLogger, err := logging.New(os.Getenv("LOG_LEVEL"))

	if err != nil {
		panic(fmt.Errorf("error %s creating serviceLogger with level %s", err, logLevel))
	}

	// parse database configuration from the environment
	databaseConfig := database.PostgresDatabaseConfig{
		DatabaseName:          os.Getenv("DATABASE_NAME"),
		DatabaseEndpointURL:   os.Getenv("DATABASE_ENDPOINT_URL"),
		DatabaseUsername:      os.Getenv("DATABASE_USERNAME"),
		DatabasePassword:      os.Getenv("DATABASE_PASSWORD"),
		SSLEnabled:            os.Getenv("DATABASE_SSL_ENABLED") == "true",
		QueryLoggingEnabled:   os.Getenv("DATABASE_QUERY_LOGGING_ENABLED") == "true",
		RunDatabaseMigrations: os.Getenv("RUN_DATABASE_MIGRATIONS") == "true",
		Logger:                &serviceLogger,
	}

	serviceLogger.Trace().Msgf("loaded databaseClient config %+v", databaseConfig)

	// parse api config from the environment
	apiConfig := service.APIConfig{
		APIPort:        os.Getenv("API_PORT"),
		DatabaseConfig: databaseConfig,
		ServiceLogger:  &serviceLogger,
		UserCookies:    api.UserCookies{},
	}

	serviceLogger.Debug().Msgf("loaded api config %+v",
		apiConfig)

	apiService, err := service.NewAPIService(serviceCtx, apiConfig)

	if err != nil {
		panic(err)
	}

	serviceLogger.Debug().Msg("api server starting")
	err = apiService.Run()

	if err != nil {
		serviceLogger.Error().Msgf("service exited with error %s", err)
	}
}
