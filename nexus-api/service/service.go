package service

import (
	"context"
	"fmt"
	"net/http"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/clients/database/schemas/postgres/migrations"
	"nexus-api/logging"
	"time"

	"github.com/gorilla/mux"
)

type APIConfig struct {
	ServiceLogger  *logging.ServiceLogger
	DatabaseConfig database.PostgresDatabaseConfig
	APIPort        string
	// #TODO refactor to storing cookies in database
	UserCookies api.UserCookies
}

type APIService struct {
	server         *http.Server
	Ctx            context.Context
	Config         APIConfig
	DatabaseClient *database.PostgresClient
	*logging.ServiceLogger
	// #TODO refactor to storing cookies in database
	UserCookies api.UserCookies
}

func (as *APIService) Run() error {
	// run api service listening on the configured port
	return as.server.ListenAndServe()
}

func NewAPIService(ctx context.Context, config APIConfig) (APIService, error) {
	nexusAPI := APIService{}

	// create database client
	databaseClient, err := database.NewPostgresClient(config.DatabaseConfig)

	if err != nil {
		return nexusAPI, fmt.Errorf("error %s creating database client with %+v", err, config.DatabaseConfig)
	}

	// run migrations based on configuration
	if config.DatabaseConfig.RunDatabaseMigrations {
		go func() {
			for {
				ranMigrations, err := database.Migrate(ctx, databaseClient.DB, *migrations.Migrations, config.ServiceLogger)

				if err != nil {
					config.ServiceLogger.Error().Msgf("error %s running migrations %+v, will retry in 3 seconds", err, migrations.Migrations)

					time.Sleep(3 * time.Second)

					continue
				}

				config.ServiceLogger.Info().Msgf("successfully ran migrations %+v", ranMigrations)

				return
			}
		}()
	}

	// setup api request router
	router := mux.NewRouter()

	// setup handler functions to run whenever a specific api endpoint is called
	router.HandleFunc("/healthcheck", CorsMiddleware(CreateHealthCheckHandler(&databaseClient)))
	router.HandleFunc("/login", CorsMiddleware(CreateLoginHandler(&nexusAPI)))
	router.HandleFunc("/change-password", CorsMiddleware(AuthMiddleware(CreateChangePasswordHandler(&nexusAPI), config.UserCookies)))

	router.HandleFunc("/settings", CorsMiddleware(AuthMiddleware(SettingsHandler, config.UserCookies)))   // Protect the settings route
	router.HandleFunc("/home", CorsMiddleware(AuthMiddleware(HomeHandler, config.UserCookies)))           // Protect the home route
	router.HandleFunc("/solar", CorsMiddleware(AuthMiddleware(SolarHandler, config.UserCookies)))         //protects solar route
	router.HandleFunc("/locations", CorsMiddleware(AuthMiddleware(LocationsHandler, config.UserCookies))) //p protects location route

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.APIPort),
		Handler: router,
	}

	nexusAPI = APIService{
		server:         server,
		Ctx:            ctx,
		Config:         config,
		DatabaseClient: &databaseClient,
		ServiceLogger:  config.ServiceLogger,
		UserCookies:    config.UserCookies,
	}

	nexusAPI.Trace().Msg(fmt.Sprintf("created nexus api  %+v", nexusAPI))

	return nexusAPI, nil
}
