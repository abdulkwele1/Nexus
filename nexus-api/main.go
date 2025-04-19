package main

import (
	"context"
	"fmt"
	"nexus-api/clients/database"
	mqttclient "nexus-api/clients/mqtt"
	"nexus-api/logging"
	"nexus-api/sdk"
	"nexus-api/service"

	"os"
	"strings"
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

	// Check if MQTT is enabled
	enableMQTT := os.Getenv("ENABLE_MQTT") == "true"
	var mqttClient *mqttclient.MQTTClient
	var sdkClient *sdk.NexusClient

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
		sdkClient, err = sdk.NewClient(sdkConfig)
		if err != nil {
			panic(err)
		}
		serviceLogger.Info().Msg("SDK client initialized successfully")
	}

	if enableMQTT {
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
		var err error
		mqttClient, err = mqttclient.NewMQTTClient(mqttConfig)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to initialize MQTT client")
			os.Exit(1)
		}
		defer mqttClient.Disconnect()
		serviceLogger.Info().Msg("MQTT client initialized successfully")
	} else {
		serviceLogger.Info().Msg("MQTT is disabled, skipping MQTT client initialization")
	}

	if enableMQTT {
		// Subscribe to a single MQTT topic
		mqttTopics := os.Getenv("MQTT_TOPICS")
		if mqttTopics == "" {
			// Default topic if none specified
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
		}
	}

	// parse api config from the environment
	apiConfig := service.APIConfig{
		APIPort:        os.Getenv("API_PORT"),
		DatabaseConfig: databaseConfig,
		ServiceLogger:  &serviceLogger,
	}

	serviceLogger.Debug().Msgf("loaded api config %+v",
		apiConfig)

	apiService, err := service.NewAPIService(serviceCtx, apiConfig)

	if err != nil {
		panic(err)
	}

	serviceLogger.Debug().Msg("api server starting")
	err = apiService.Run(serviceCtx)

	if err != nil {
		serviceLogger.Error().Msgf("service exited with error %s", err)
	}
}
