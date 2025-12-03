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
	router.HandleFunc("/refresh-session", CorsMiddleware(AuthMiddleware(CreateSessionRefreshHandler(&nexusAPI), &nexusAPI)))
	router.HandleFunc("/change-password", CorsMiddleware(AuthMiddleware(CreateChangePasswordHandler(&nexusAPI), &nexusAPI)))

	router.HandleFunc("/settings", CorsMiddleware(AuthMiddleware(CreateSettingsHandler(&nexusAPI), &nexusAPI)))
	router.HandleFunc("/home", CorsMiddleware(AuthMiddleware(CreateHomeHandler(&nexusAPI), &nexusAPI)))
	router.HandleFunc("/solar", CorsMiddleware(AuthMiddleware(CreateSolarHandler(&nexusAPI), &nexusAPI)))
	router.HandleFunc("/locations", CorsMiddleware(AuthMiddleware(CreateLocationsHandler(&nexusAPI), &nexusAPI)))

	//new routes for Kwh logging + retrieval
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateGetPanelYieldDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/panels/{panel_id}/yield_data", CorsMiddleware(AuthMiddleware(CreateSetPanelYieldDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	//new routes for Consumption logging + retrieval
	router.HandleFunc("/panels/{panel_id}/consumption_data", CorsMiddleware(AuthMiddleware(CreateGetPanelConsumptionDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/panels/{panel_id}/consumption_data", CorsMiddleware(AuthMiddleware(CreateSetPanelConsumptionDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	//new routes for SensorMoisture logging + retrieval
	router.HandleFunc("/sensors/{sensor_id}/moisture_data", CorsMiddleware(AuthMiddleware(CreateGetSensorMoistureDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/sensors/{sensor_id}/moisture_data", CorsMiddleware(AuthMiddleware(CreateSetSensorMoistureDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	//new routes for SensorTemperature logging + retrieval
	router.HandleFunc("/sensors/{sensor_id}/temperature_data", CorsMiddleware(AuthMiddleware(CreateGetSensorTemperatureDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/sensors/{sensor_id}/temperature_data", CorsMiddleware(AuthMiddleware(CreateSetSensorTemperatureDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	// Route to get all sensors
	router.HandleFunc("/sensors", CorsMiddleware(AuthMiddleware(CreateGetAllSensorsHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)

	// Route to add a new sensor
	router.HandleFunc("/sensors", CorsMiddleware(AuthMiddleware(CreateAddSensorHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost, http.MethodOptions)

	// Route to delete a sensor
	router.HandleFunc("/sensors/{sensor_id}", CorsMiddleware(AuthMiddleware(CreateDeleteSensorHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodDelete, http.MethodOptions)

	// Drone image routes
	router.HandleFunc("/drone_images", CorsMiddleware(AuthMiddleware(CreateGetDroneImagesHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/drone_images", CorsMiddleware(AuthMiddleware(CreateUploadDroneImagesHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)
	router.HandleFunc("/drone_images/{image_id}", CorsMiddleware(AuthMiddleware(CreateGetDroneImageHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/drone_images/{image_id}", CorsMiddleware(AuthMiddleware(CreateDeleteDroneImageHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodDelete)
	router.HandleFunc("/drone_images/{image_id}/content", CorsMiddleware(AuthMiddleware(CreateGetDroneImageContentHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)

	// Battery data routes
	router.HandleFunc("/sensors/{sensor_id}/battery_data", CorsMiddleware(AuthMiddleware(CreateGetSensorBatteryDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/sensors/{sensor_id}/battery_data", CorsMiddleware(AuthMiddleware(CreateSetSensorBatteryDataHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost)

	// Admin routes
	router.HandleFunc("/admin/users", CorsMiddleware(AdminMiddleware(CreateGetAllUsersHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/admin/users", CorsMiddleware(AdminMiddleware(CreateCreateUserHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/admin/users/check/{username}", CorsMiddleware(AdminMiddleware(CreateCheckUsernameHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/admin/users/{username}", CorsMiddleware(AdminMiddleware(CreateUpdateUserRoleHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodPatch, http.MethodOptions)
	router.HandleFunc("/admin/users/{username}/remove-admin", CorsMiddleware(AdminMiddleware(CreateRemoveAdminHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodDelete, http.MethodOptions)
	router.HandleFunc("/admin/users/{username}", CorsMiddleware(AdminMiddleware(CreateDeleteUserHandler(&nexusAPI), &nexusAPI))).Methods(http.MethodDelete, http.MethodOptions)

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
