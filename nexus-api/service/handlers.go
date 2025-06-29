package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/password"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func CreateHealthCheckHandler(databaseClient *database.PostgresClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var combinedErrors error

		databaseErr := databaseClient.HealthCheck()

		if databaseErr != nil {
			errMsg := fmt.Errorf("error %s unable to connect to database", databaseErr)
			combinedErrors = errors.Join(combinedErrors, errMsg)
		}

		if combinedErrors != nil {
			w.WriteHeader(http.StatusInternalServerError)

			w.Write([]byte(combinedErrors.Error()))

			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("nexus api is healthy"))
	}
}

func CreateLoginHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request api.LoginRequest
		err := json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			apiService.Debug().Msgf("error %s parsing %+v", err, request)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		apiService.Debug().Msgf("login username %s, password %s\n", request.Username, request.Password)

		// Check if username doesn't exist in our system
		loginAuthentication, err := database.GetLoginAuthenticationByUserName(apiService.Ctx, apiService.DatabaseClient.DB, request.Username)
		if err != nil {
			if errors.Is(err, database.ErrorNoLoginAuthenticationForUsername) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "User not found"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Check the password
		match := password.CheckPasswordHash(request.Password, loginAuthentication.PasswordHash)

		if !match {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Generate a session cookie
		response := api.LoginResponse{
			RedirectURL: "/",
			Match:       true,
			Cookie:      uuid.NewString(),
		}

		apiService.Trace().Msgf("password hash for user %s in our system is %s", request.Username, loginAuthentication.PasswordHash)

		// Set the cookie with an expiration time
		expiration := time.Now().Add(3 * 24 * time.Hour)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    response.Cookie,
			Path:     "/",
			Expires:  expiration,
			MaxAge:   3600,  // 1 hour
			HttpOnly: true,  // Optional: helps mitigate XSS
			Secure:   false, // Set to true if serving over HTTPS
		})

		// upsert cookie to database
		loginCookie := database.LoginCookie{
			UserName:   request.Username,
			Cookie:     response.Cookie,
			Expiration: expiration,
		}

		err = loginCookie.Save(r.Context(), apiService.DatabaseClient.DB)

		if err != nil {
			apiService.Error().Msgf("error %s saving login cookie %+v to database", err, loginCookie)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{}{})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateLogoutHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)

		userName, ok := rawUsername.(string)

		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: http.StatusText(http.StatusInternalServerError)})
			return
		}

		err := database.DeleteCookieForUserName(r.Context(), userName, apiService.DatabaseClient.DB)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: http.StatusText(http.StatusInternalServerError)})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(struct{}{})
	}
}

func CreateChangePasswordHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := api.ChangePasswordRequest{}
		err := json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		username, ok := r.Context().Value("username").(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		apiService.Debug().Msgf("User attempting to change password: %s", username)

		loginAuthentication, err := database.GetLoginAuthenticationByUserName(apiService.Ctx, apiService.DatabaseClient.DB, username)
		if err != nil {
			apiService.Error().Msgf("error retrieving login authentication for user %s: %s", username, err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if the current password is correct
		if !password.CheckPasswordHash(request.CurrentPassword, loginAuthentication.PasswordHash) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Current password is incorrect"})
			return
		}

		// Hash the new password
		newPasswordHash, err := password.HashPassword(request.NewPassword)
		if err != nil {
			apiService.Error().Msgf("error hashing new password for user %s: %s", username, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Update the password hash in the database
		loginAuthentication.PasswordHash = newPasswordHash
		err = loginAuthentication.Update(apiService.Ctx, apiService.DatabaseClient.DB)
		if err != nil {
			apiService.Error().Msgf("error updating password for user %s: %s", username, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Optionally, you can clear the old password hash variable if needed
		// (not strictly necessary in this context since it's being replaced)
		loginAuthentication.PasswordHash = ""

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct{}{})
	}
}

func LocationsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
}

func SolarHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Settings page - only accessible with a valid cookie! User: %s", username)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	fmt.Fprintf(w, "Home page - only accessible with a valid cookie! User: %s", username)
}

func CreateGetPanelYieldDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		panelIDRaw := vars["panel_id"]

		// Parse and validate panelID as an integer
		panelID, err := strconv.Atoi(panelIDRaw)
		if err != nil {
			apiService.Error().Msgf("Invalid panel_id: %s", panelIDRaw)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid panel_id"})
			return
		}

		// Retrieve data for the panelID
		data, err := database.GetYieldDataForPanelID(r.Context(), apiService.DatabaseClient.DB, panelID)
		if err != nil {
			if errors.Is(err, database.ErrorNoSolarPanelYieldData) {
				apiService.Debug().Msgf("No data found for panel_id: %d", panelID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No data found"})
				return
			}

			apiService.Error().Msgf("Error retrieving data for panel_id: %d, error: %s", panelID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Convert database data to GetPanelYieldDataResponse
		var response api.GetPanelYieldDataResponse
		for _, d := range data {
			response.YieldData = append(response.YieldData, api.YieldData{
				Date:     d.Date,
				KwhYield: d.KwHYield,
			})
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateSetPanelYieldDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		panelIDRaw := vars["panel_id"]

		// Parse and validate panelID as an integer
		panelID, err := strconv.Atoi(panelIDRaw)
		if err != nil {
			apiService.Error().Msgf("Invalid panel_id: %s", panelIDRaw)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid panel_id"})
			return
		}

		request := api.SetPanelYieldDataResponse{}
		err = json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		// Iterate over each YieldData item and save to the database
		for _, yieldData := range request.YieldData {
			solarPanelData := database.SolarPanelYieldData{
				Date:     yieldData.Date,
				KwHYield: yieldData.KwhYield,
				PanelID:  panelID,
			}

			err := solarPanelData.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Failed to save yield data for panel_id: %d, data: %+v, error: %s", panelID, solarPanelData, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to save yield data"})
				return
			}
		}

		// Send success response
		apiService.Trace().Msgf("Successfully saved yield data for panel_id: %d", panelID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Yield data saved successfully"})
	}
}

func CreateGetPanelConsumptionDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		panelIDRaw := vars["panel_id"]

		// Parse and validate panelID as an integer
		panelID, err := strconv.Atoi(panelIDRaw)
		if err != nil {
			apiService.Error().Msgf("Invalid panel_id: %s", panelIDRaw)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid panel_id"})
			return
		}

		// Retrieve data for the panelID
		data, err := database.GetConsumptionDataForPanelID(r.Context(), apiService.DatabaseClient.DB, panelID)
		if err != nil {
			if errors.Is(err, database.ErrorNoSolarPanelConsumptionData) {
				apiService.Debug().Msgf("No data found for panel_id: %d", panelID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No data found"})
				return
			}

			apiService.Error().Msgf("Error retrieving data for panel_id: %d, error: %s", panelID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Convert database data to GetPanelConsumptionDataResponse
		var response api.GetPanelConsumptionDataResponse
		for _, d := range data {
			response.ConsumptionData = append(response.ConsumptionData, api.ConsumptionData{
				Date:        d.Date,
				CapacityKwh: d.CapacityKwh,
				ConsumedKwh: d.ConsumedKwh,
			})
		}
		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateSetPanelConsumptionDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		panelIDRaw := vars["panel_id"]

		// Parse and validate panelID as an integer
		panelID, err := strconv.Atoi(panelIDRaw)
		if err != nil {
			apiService.Error().Msgf("Invalid panel_id: %s", panelIDRaw)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid panel_id"})
			return
		}

		request := api.SetPanelConsumptionDataResponse{}
		err = json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		// Iterate over each ConsumptionData item and save to the database
		for _, ConsumptionData := range request.ConsumptionData {
			solarPanelData := database.SolarPanelConsumptionData{
				Date:        ConsumptionData.Date,
				CapacityKwh: ConsumptionData.CapacityKwh,
				ConsumedKwh: ConsumptionData.ConsumedKwh,
				PanelID:     panelID,
			}

			err := solarPanelData.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Failed to save Consumption data for panel_id: %d, data: %+v, error: %s", panelID, solarPanelData, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to save Consumption data"})
				return
			}
		}

		// Send success response
		apiService.Trace().Msgf("Successfully saved consumption data for panel_id: %d", panelID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Consumption data saved successfully"})
	}
}

func CreateGetSensorMoistureDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]
		log.Info().Msgf("[handlers.go] Received request for moisture data for sensor_id: %s", sensorID)

		// Retrieve data for the sensorID
		data, err := database.GetSensorMoistureDataForSensorID(r.Context(), apiService.DatabaseClient.DB, sensorID)
		if err != nil {
			if errors.Is(err, database.ErrorNoSensorMoistureData) {
				apiService.Debug().Msgf("No data found for sensor_id: %s", sensorID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No data found"})
				return
			}

			apiService.Error().Msgf("Error retrieving data for sensor_id: %s, error: %s", sensorID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Convert database data to GetSensorMoistureDataResponse
		var response api.GetSensorMoistureDataResponse
		for _, d := range data {
			response.SensorMoistureData = append(response.SensorMoistureData, api.SensorMoistureData{
				ID:           d.ID,
				SensorID:     d.SensorID,
				Date:         d.Date,
				SoilMoisture: d.SoilMoisture,
			})
		}
		apiService.Debug().Msgf("Sending back %d moisture data records for sensor_id: %s", len(response.SensorMoistureData), sensorID)

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateSetSensorMoistureDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]

		request := api.SetSensorMoistureDataResponse{}
		err := json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		// Ensure sensor exists before saving data
		err = database.EnsureSensorExists(r.Context(), apiService.DatabaseClient.DB, sensorID, sensorID)
		if err != nil {
			apiService.Error().Msgf("Failed to ensure sensor exists for sensor_id: %s, error: %s", sensorID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to ensure sensor exists"})
			return
		}

		// Iterate over each SensorMoistureData item and save to the database
		for _, sensorMoistureData := range request.SensorMoistureData {
			moistureData := database.SensorMoistureData{
				SensorID:     sensorID,
				Date:         sensorMoistureData.Date,
				SoilMoisture: sensorMoistureData.SoilMoisture,
			}

			err = moistureData.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Failed to save moisture data for sensor_id: %s, data: %+v, error: %s", sensorID, moistureData, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to save moisture data"})
				return
			}
		}

		// Send success response
		apiService.Trace().Msgf("Successfully saved moisture data for sensor_id: %s", sensorID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Moisture data saved successfully"})
	}
}

func CreateGetSensorTemperatureDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]

		// Retrieve data for the sensorID
		data, err := database.GetSensorTemperatureDataForSensorID(r.Context(), apiService.DatabaseClient.DB, sensorID)
		if err != nil {
			if errors.Is(err, database.ErrorNoSensorTemperatureData) {
				apiService.Debug().Msgf("No data found for sensor_id: %s", sensorID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No data found"})
				return
			}

			apiService.Error().Msgf("Error retrieving data for sensor_id: %s, error: %s", sensorID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Convert database data to GetSensorTemperatureDataResponse
		var response api.GetSensorTemperatureDataResponse
		for _, d := range data {
			response.SensorTemperatureData = append(response.SensorTemperatureData, api.SensorTemperatureData{
				ID:              d.ID,
				SensorID:        d.SensorID,
				Date:            d.Date,
				SoilTemperature: d.SoilTemperature,
			})
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateSetSensorTemperatureDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]

		request := api.SetSensorTemperatureDataResponse{}
		err := json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request"})
			return
		}

		// Ensure sensor exists before saving data
		err = database.EnsureSensorExists(r.Context(), apiService.DatabaseClient.DB, sensorID, sensorID)
		if err != nil {
			apiService.Error().Msgf("Failed to ensure sensor exists for sensor_id: %s, error: %s", sensorID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to ensure sensor exists"})
			return
		}

		// Iterate over each SensorTemperatureData item and save to the database
		for _, sensorTemperatureData := range request.SensorTemperatureData {
			temperatureData := database.SensorTemperatureData{
				SensorID:        sensorID,
				Date:            sensorTemperatureData.Date,
				SoilTemperature: sensorTemperatureData.SoilTemperature,
			}

			err = temperatureData.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Failed to save temperature data for sensor_id: %s, data: %+v, error: %s", sensorID, temperatureData, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to save temperature data"})
				return
			}
		}

		// Send success response
		apiService.Trace().Msgf("Successfully saved temperature data for sensor_id: %s", sensorID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Temperature data saved successfully"})
	}
}

func CreateGetAllSensorsHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			http.Error(w, "failed to get username from context", http.StatusUnauthorized)
			return
		}

		sensors, err := apiService.DatabaseClient.GetAllSensors(ctx, username)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to get all sensors from database")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(sensors) == 0 {
			log.Ctx(ctx).Debug().Msg("no sensors found in database")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{}) // Return empty array instead of erroring
			return
		}

		log.Ctx(ctx).Debug().Int("count", len(sensors)).Msg("sending back sensors")
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(sensors)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to encode sensors to json")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
