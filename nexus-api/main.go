package main

import (
	"context"

	"fmt"

	"net/http"

	"nexus-api/api"

	"nexus-api/clients/database"

	mqttclient "nexus-api/clients/mqtt"

	"nexus-api/logging"

	"nexus-api/sdk"

	"nexus-api/service"

	"os"

	"strings"

	"sync"

	"time"
)

var (
	serviceCtx = context.Background()
)

func main() {

	// setup logger

	logLevel := os.Getenv("LOG_LEVEL")

	serviceLogger, err := logging.New(os.Getenv("LOG_LEVEL"))

	if err != nil {

		panic(fmt.Errorf("error %s creating serviceLogger with level %s", err, logLevel))

	}

	// parse database configuration from the environment

	databaseConfig := database.PostgresDatabaseConfig{

		DatabaseName: os.Getenv("DATABASE_NAME"),

		DatabaseEndpointURL: os.Getenv("DATABASE_ENDPOINT_URL"),

		DatabaseUsername: os.Getenv("DATABASE_USERNAME"),

		DatabasePassword: os.Getenv("DATABASE_PASSWORD"),

		SSLEnabled: os.Getenv("DATABASE_SSL_ENABLED") == "true",

		QueryLoggingEnabled: os.Getenv("DATABASE_QUERY_LOGGING_ENABLED") == "true",

		RunDatabaseMigrations: os.Getenv("RUN_DATABASE_MIGRATIONS") == "true",

		Logger: &serviceLogger,
	}

	serviceLogger.Trace().Msgf("loaded databaseClient config %+v", databaseConfig)

	// --- Initialize API Service (runs migrations synchronously) ---

	apiConfig := service.APIConfig{

		APIPort: os.Getenv("API_PORT"),

		DatabaseConfig: databaseConfig,

		ServiceLogger: &serviceLogger,
	}

	serviceLogger.Debug().Msgf("loaded api config %+v", apiConfig)

	apiService, err := service.NewAPIService(serviceCtx, apiConfig)

	if err != nil {

		// If NewAPIService fails (e.g., migrations failed), we cannot continue

		panic(fmt.Errorf("failed to create API service (migrations might have failed): %w", err))

	}

	serviceLogger.Info().Msg("API Service initialized successfully (migrations complete if enabled).")

	// --- Start API Server in Goroutine ---

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {

		defer wg.Done()

		serviceLogger.Info().Msg("Starting API server listener...")

		err := apiService.Run(serviceCtx)

		if err != nil && err != http.ErrServerClosed {

			serviceLogger.Error().Err(err).Msg("API server listener exited with error")

		} else {

			serviceLogger.Info().Msg("API server listener stopped gracefully")

		}

	}()

	// --- Initialize SDK and MQTT Clients (if enabled and API is healthy) ---

	enableMQTT := os.Getenv("ENABLE_MQTT") == "true"

	var mqttClient *mqttclient.MQTTClient

	// var sdkClient *sdk.NexusClient // sdkClient defined below if needed

	if enableMQTT {

		serviceLogger.Info().Msg("API is healthy, initializing MQTT and SDK clients...")

		// Setup SDK client config from environment

		sdkConfig := sdk.SDKConfig{

			NexusAPIEndpoint: os.Getenv("NEXUS_API_URL"), // Ensure this uses the correct reachable URL

			UserName: os.Getenv("NEXUS_API_USERNAME"),

			Password: os.Getenv("NEXUS_API_PASSWORD"),

			Logger: &serviceLogger,
		}

		serviceLogger.Info().Msgf("Initialized SDK client config - Endpoint: %s, Username: %s", sdkConfig.NexusAPIEndpoint, sdkConfig.UserName)

		// Initialize SDK client

		sdkClient, err := sdk.NewClient(sdkConfig)

		if err != nil {

			// Changed panic to error log + potentially continue/exit

			serviceLogger.Error().Err(err).Msg("Failed to initialize SDK client")

			// Consider os.Exit(1) here if SDK is critical

		} else {

			serviceLogger.Info().Msg("SDK client initialized successfully")

			// Login the SDK client *after* API is healthy, retrying until successful

			loginParams := api.LoginRequest{

				Username: sdkConfig.UserName,

				Password: sdkConfig.Password,
			}

			loginRetryDelay := 5 * time.Second

			for { // Retry loop for SDK login

				_, err = sdkClient.Login(serviceCtx, loginParams)

				if err == nil {

					serviceLogger.Info().Msg("SDK client logged in successfully")

					break // Exit loop on successful login

				}

				// Log error and wait before retrying

				serviceLogger.Error().Err(err).Msgf("Failed to login SDK client, retrying in %v...", loginRetryDelay)

				// Consider adding a maximum retry count or backoff strategy for robustness

				time.Sleep(loginRetryDelay)

			}

			// Continue ONLY after successful login...

			// parse MQTT configuration from the environment

			mqttConfig := mqttclient.MQTTConfig{

				BrokerURL: os.Getenv("MQTT_BROKER_URL"),

				ClientID: os.Getenv("MQTT_CLIENT_ID"),

				Username: os.Getenv("MQTT_USERNAME"),

				Password: os.Getenv("MQTT_PASSWORD"),

				CleanSession: os.Getenv("MQTT_CLEAN_SESSION") != "false",

				AutoReconnect: os.Getenv("MQTT_AUTO_RECONNECT") != "false",

				Logger: &serviceLogger,

				SDKClient: sdkClient, // Pass the initialized and logged-in SDK client

			}

			serviceLogger.Trace().Msgf("loaded MQTT client config %+v", mqttConfig)

			// Initialize MQTT client

			mqttClient, err = mqttclient.NewMQTTClient(mqttConfig)

			if err != nil {

				serviceLogger.Error().Err(err).Msg("Failed to initialize MQTT client")

				// os.Exit(1)

			} else {

				defer mqttClient.Disconnect()

				serviceLogger.Info().Msg("MQTT client initialized successfully")

				// Subscribe to MQTT topics (run in a separate goroutine)

				wg.Add(1)

				go func() {

					defer wg.Done()

					mqttTopics := os.Getenv("MQTT_TOPICS")

					if mqttTopics == "" {
						mqttTopics = "/device_sensor_data/444574498032128/+/+/+/+" // Subscribe to your organization's sensor data
					}

					topic := strings.Split(strings.TrimSpace(mqttTopics), ",")[0]

					topic = strings.TrimSpace(topic)

					if topic != "" {

						err := mqttClient.Subscribe(serviceCtx, topic, 1, mqttClient.HandleMessage)

						if err != nil {

							serviceLogger.Error().Err(err).Msgf("Failed to subscribe to topic: %s", topic)

						} else {

							serviceLogger.Info().Msgf("Subscribed to topic: %s", topic)

						}

					} else {

						serviceLogger.Warn().Msg("No MQTT topic specified to subscribe to.")

					}

				}()

			}

		}

	} else {

		serviceLogger.Info().Msg("MQTT is disabled, skipping MQTT/SDK client initialization")

	}

	serviceLogger.Info().Msg("Initialization complete. Waiting for services to finish...")

	wg.Wait()

	serviceLogger.Info().Msg("Application shutting down.")

}
