package service

import (
	"context"
	"fmt"
	"net/http"
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
}

type APIService struct {
	server         *http.Server
	Ctx            context.Context
	Config         APIConfig
	DatabaseClient *database.PostgresClient
	*logging.ServiceLogger
}

func (as *APIService) Run(ctx context.Context) error {
	// run background routine to delete any expired cookies
	go func() {
		as.ExpireCookies(ctx)
	}()
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
	router.HandleFunc("/logout", CorsMiddleware(AuthMiddleware(CreateLogoutHandler(&nexusAPI), &nexusAPI)))
	router.HandleFunc("/change-password", CorsMiddleware(AuthMiddleware(CreateChangePasswordHandler(&nexusAPI), &nexusAPI)))

	router.HandleFunc("/settings", CorsMiddleware(AuthMiddleware(SettingsHandler, &nexusAPI)))   // Protect the settings route
	router.HandleFunc("/home", CorsMiddleware(AuthMiddleware(HomeHandler, &nexusAPI)))           // Protect the home route
	router.HandleFunc("/solar", CorsMiddleware(AuthMiddleware(SolarHandler, &nexusAPI)))         //protects solar route
	router.HandleFunc("/locations", CorsMiddleware(AuthMiddleware(LocationsHandler, &nexusAPI))) //p protects location route

	//new routes for Kwh logging + retrieval
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateGetPanelYieldDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateSetPanelYieldDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	//new routes for Consumption logging + retrieval
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateGetPanelConsumptionDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateSetPanelConsumptionDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

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
	}

	nexusAPI.Trace().Msg(fmt.Sprintf("created nexus api  %+v", nexusAPI))

	return nexusAPI, nil
}

func (as *APIService) ExpireCookies(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			as.Trace().Msgf("ExpireCookies routine running %+v", t)
			err := database.DeleteExpiredCookies(ctx, time.Now(), as.DatabaseClient.DB)

			if err != nil {
				as.Error().Msgf("error %s deleting expired cookies", err)
			}
		}
	}
}
