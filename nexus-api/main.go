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

	// --- Start API Service First ---
	apiConfig := service.APIConfig{
		APIPort:        os.Getenv("API_PORT"),
		DatabaseConfig: databaseConfig,
		ServiceLogger:  &serviceLogger,
	}
	serviceLogger.Debug().Msgf("loaded api config %+v", apiConfig)

	apiService, err := service.NewAPIService(serviceCtx, apiConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create API service: %w", err))
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		serviceLogger.Info().Msg("Starting API server...")
		err := apiService.Run(serviceCtx)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("API service exited with error")
		} else {
			serviceLogger.Info().Msg("API service stopped gracefully")
		}
	}()

	// --- Wait for API server to be healthy before proceeding ---
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080" // Default port if not set
	}
	healthCheckURL := fmt.Sprintf("http://localhost:%s/healthcheck", apiPort)
	maxWait := 30 * time.Second // Max time to wait for health check
	checkInterval := 200 * time.Millisecond
	startTime := time.Now()

	serviceLogger.Info().Msgf("Waiting for API server at %s to be healthy...", healthCheckURL)
	for time.Since(startTime) < maxWait {
		resp, err := http.Get(healthCheckURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			serviceLogger.Info().Msg("API server is healthy.")
			resp.Body.Close() // Important to close the body
			break             // Exit loop on success
		}
		if resp != nil {
			resp.Body.Close() // Close body even on non-200 status
		}
		if time.Since(startTime)+checkInterval >= maxWait {
			serviceLogger.Error().Err(err).Int("status", resp.StatusCode).Msgf("API server health check timed out after %v", maxWait)
			os.Exit(1)
		}
		time.Sleep(checkInterval)
	}

	// --- Initialize SDK and MQTT Clients (if enabled) ---
	enableMQTT := os.Getenv("ENABLE_MQTT") == "true"
	var mqttClient *mqttclient.MQTTClient

	if enableMQTT {
		// Setup SDK client config from environment
		sdkConfig := sdk.SDKConfig{
			NexusAPIEndpoint: os.Getenv("NEXUS_API_URL"),
			UserName:         os.Getenv("NEXUS_API_USERNAME"),
			Password:         os.Getenv("NEXUS_API_PASSWORD"),
			Logger:           &serviceLogger,
		}
		serviceLogger.Trace().Msgf("loaded SDK client config %+v", sdkConfig)

		// Initialize SDK client
		sdkClient, err := sdk.NewClient(sdkConfig)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to initialize SDK client")
<<<<<<< Updated upstream
			os.Exit(1)
=======
			// Consider os.Exit(1) here if SDK is critical
		} else {
			serviceLogger.Info().Msg("SDK client initialized successfully")

			// Login the SDK client *after* API is healthy
			loginParams := api.LoginRequest{
				Username: sdkConfig.UserName,
				Password: sdkConfig.Password,
			}
			_, err = sdkClient.Login(serviceCtx, loginParams)
			if err != nil {
				serviceLogger.Error().Err(err).Msg("Failed to login SDK client")
				// Consider os.Exit(1) here if login is critical
			} else {
				serviceLogger.Info().Msg("SDK client logged in successfully")

				// parse MQTT configuration from the environment
				mqttConfig := mqttclient.MQTTConfig{
					BrokerURL:     os.Getenv("MQTT_BROKER_URL"),
					ClientID:      os.Getenv("MQTT_CLIENT_ID"),
					Username:      os.Getenv("MQTT_USERNAME"),
					Password:      os.Getenv("MQTT_PASSWORD"),
					CleanSession:  os.Getenv("MQTT_CLEAN_SESSION") != "false",
					AutoReconnect: os.Getenv("MQTT_AUTO_RECONNECT") != "false",
					Logger:        &serviceLogger,
					SDKClient:     sdkClient, // Pass the initialized and logged-in SDK client
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
							mqttTopics = "/device_sensor_data/444574498032128/+/+/+/+" // Default topic
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
>>>>>>> Stashed changes
		}
		serviceLogger.Info().Msg("SDK client initialized successfully")

		// Login the SDK client
		loginParams := api.LoginRequest{
			Username: sdkConfig.UserName,
			Password: sdkConfig.Password,
		}
		_, err = sdkClient.Login(serviceCtx, loginParams)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to login SDK client")
			os.Exit(1)
		}
		serviceLogger.Info().Msg("SDK client logged in successfully")

		// parse MQTT configuration from the environment
		mqttConfig := mqttclient.MQTTConfig{
			BrokerURL:     os.Getenv("MQTT_BROKER_URL"),
			ClientID:      os.Getenv("MQTT_CLIENT_ID"),
			Username:      os.Getenv("MQTT_USERNAME"),
			Password:      os.Getenv("MQTT_PASSWORD"),
			CleanSession:  os.Getenv("MQTT_CLEAN_SESSION") != "false",
			AutoReconnect: os.Getenv("MQTT_AUTO_RECONNECT") != "false",
			Logger:        &serviceLogger,
			SDKClient:     sdkClient,
		}
		serviceLogger.Trace().Msgf("loaded MQTT client config %+v", mqttConfig)

		// Initialize MQTT client
		mqttClient, err = mqttclient.NewMQTTClient(mqttConfig)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to initialize MQTT client")
			os.Exit(1)
		}
		defer mqttClient.Disconnect()
		serviceLogger.Info().Msg("MQTT client initialized successfully")

		// Subscribe to MQTT topics (run in a separate goroutine)
		wg.Add(1)
		go func() {
			defer wg.Done()
			mqttTopics := os.Getenv("MQTT_TOPICS")
			if mqttTopics == "" {
				mqttTopics = "/device_sensor_data/444574498032128/+/+/+/+"
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

	} else {
		serviceLogger.Info().Msg("MQTT is disabled, skipping MQTT/SDK client initialization")
	}

	serviceLogger.Info().Msg("Initialization complete. Waiting for services to finish...")
	wg.Wait()

	serviceLogger.Info().Msg("Application shutting down.")
}
