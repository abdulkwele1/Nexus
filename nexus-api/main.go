package main

import (
	"context"
	"fmt"
	"nexus-api/clients/database"
	"nexus-api/logging"
	"nexus-api/service"

	"os"
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

	func AddPanelData(w http.ResponseWriter, r *http.Request) {
		panelID := mux.Vars(r)["id"] // Panel ID from the URL
		var input struct {
			Date       string  `json:"date"`
			Production float64 `json:"production"`
		}
	
		// Decode JSON request body
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
	
		// Insert data into the database
		query := `INSERT INTO panel_data (panel_id, date, production) VALUES ($1, $2, $3)`
		_, err := db.Exec(query, panelID, input.Date, input.Production)
		if err != nil {
			http.Error(w, "Failed to add data", http.StatusInternalServerError)
			return
		}
	
		w.WriteHeader(http.StatusCreated)
	}
	
	func GetPanelData(w http.ResponseWriter, r *http.Request) {
		panelID := mux.Vars(r)["id"]
		rows, err := db.Query(`SELECT date, production FROM panel_data WHERE panel_id = $1`, panelID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
	
		var data []struct {
			Date       time.Time `json:"date"`
			Production float64   `json:"production"`
		}
	
		for rows.Next() {
			var d struct {
				Date       time.Time `json:"date"`
				Production float64   `json:"production"`
			}
			if err := rows.Scan(&d.Date, &d.Production); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data = append(data, d)
		}
	
		json.NewEncoder(w).Encode(data)
	}

	func RemovePanelData(w http.ResponseWriter, r *http.Request) {
		dataID := mux.Vars(r)["data_id"] // Data ID from the URL
	
		// Delete data from the database
		query := `DELETE FROM panel_data WHERE id = $1`
		_, err := db.Exec(query, dataID)
		if err != nil {
			http.Error(w, "Failed to remove data", http.StatusInternalServerError)
			return
		}
	
		w.WriteHeader(http.StatusOK)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/panels/{id}/data", AddPanelData).Methods("POST")
	router.HandleFunc("/api/panels/{id}/data", GetPanelData).Methods("GET")
	router.HandleFunc("/api/panels/data/{data_id}", RemovePanelData).Methods("DELETE")


}
