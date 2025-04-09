package main

import (
	"context"
	"fmt"
	"nexus-api/clients/database"
	mqttclient "nexus-api/clients/mqtt"
	"nexus-api/logging"
	"nexus-api/service"

	"os"
	"strings"
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

	// Check if MQTT is enabled
	enableMQTT := os.Getenv("ENABLE_MQTT") == "true"
	var mqttClient *mqttclient.MQTTClient
	var handlers map[string]mqttclient.MQTTMessageHandler

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

		// Define message handlers for different topics
		handlers = map[string]mqttclient.MQTTMessageHandler{
			"/device_sensor_data/444574498032128/+/+/+/+": mqttclient.DefaultSensorDataHandler,
			"system/status": mqttclient.DefaultSystemStatusHandler,
		}
	} else {
		serviceLogger.Info().Msg("MQTT is disabled, skipping MQTT client initialization")
	}

	// Create a message handler
	messageHandler := mqttclient.CreateMQTTMessageHandler(serviceCtx, &serviceLogger, handlers)

	if enableMQTT {
		// Subscribe to MQTT topics
		mqttTopics := os.Getenv("MQTT_TOPICS")
		if mqttTopics == "" {
			// Default topics if none specified
			mqttTopics = "/device_sensor_data/444574498032128/+/+/+/+"
		}

		// Subscribe to each topic
		topics := strings.Split(mqttTopics, ",")
		for _, topic := range topics {
			topic = strings.TrimSpace(topic)
			if topic != "" {
				err := mqttClient.Subscribe(serviceCtx, topic, 1, messageHandler)
				if err != nil {
					serviceLogger.Error().Err(err).Msgf("Failed to subscribe to topic: %s", topic)
				} else {
					serviceLogger.Info().Msgf("Subscribed to topic: %s", topic)
				}
			}
		}

		// Publish initial system status
		status := map[string]interface{}{
			"status":    "online",
			"version":   "1.0.0",
			"startTime": time.Now().Format(time.RFC3339),
		}
		err = mqttclient.PublishSystemStatus(serviceCtx, mqttClient, status, &serviceLogger)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to publish initial system status")
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

	// func AddPanelData(w http.ResponseWriter, r *http.Request) {
	// 	panelID := mux.Vars(r)["id"] // Panel ID from the URL
	// 	var input struct {
	// 		Date       string  `json:"date"`
	// 		Production float64 `json:"production"`
	// 	}

	// 	// Decode JSON request body
	// 	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
	// 		http.Error(w, "Invalid input", http.StatusBadRequest)
	// 		return
	// 	}

	// 	// Insert data into the database
	// 	query := `INSERT INTO panel_data (panel_id, date, production) VALUES ($1, $2, $3)`
	// 	_, err := db.Exec(query, panelID, input.Date, input.Production)
	// 	if err != nil {
	// 		http.Error(w, "Failed to add data", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	w.WriteHeader(http.StatusCreated)
	// }

	// func GetPanelData(w http.ResponseWriter, r *http.Request) {
	// 	panelID := mux.Vars(r)["id"]
	// 	rows, err := db.Query(`SELECT date, production FROM panel_data WHERE panel_id = $1`, panelID)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	defer rows.Close()

	// 	var data []struct {
	// 		Date     make  time.Time `json:"date"`
	// 		Production float64   `json:"production"`
	// 	}

	// 	for rows.Next() {
	// 		var d struct {
	// 			Date       time.Time `json:"date"`
	// 			Production float64   `json:"production"`
	// 		}
	// 		if err := rows.Scan(&d.Date, &d.Production); err != nil {
	// 			http.Error(w, err.Error(), http.StatusInternalServerError)
	// 			return
	// 		}
	// 		data = append(data, d)
	// 	}

	// 	json.NewEncoder(w).Encode(data)
	// }

	// func RemovePanelData(w http.ResponseWriter, r *http.Request) {
	// 	dataID := mux.Vars(r)["data_id"] // Data ID from the URL

	// 	// Delete data from the database
	// 	query := `DELETE FROM panel_data WHERE id = $1`
	// 	_, err := db.Exec(query, dataID)
	// 	if err != nil {
	// 		http.Error(w, "Failed to remove data", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	w.WriteHeader(http.StatusOK)
	// }

	// router := mux.NewRouter()
	// router.HandleFunc("/api/panels/{id}/data", AddPanelData).Methods("POST")
	// router.HandleFunc("/api/panels/{id}/data", GetPanelData).Methods("GET")
	// router.HandleFunc("/api/panels/data/{data_id}", RemovePanelData).Methods("DELETE")

}
