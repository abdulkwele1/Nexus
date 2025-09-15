package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/password"
	"os"
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

		// Set the cookie with an expiration time (24 hours)
		expiration := time.Now().Add(24 * time.Hour)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    response.Cookie,
			Path:     "/",
			Expires:  expiration,
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

func CreateSessionRefreshHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)

		userName, ok := rawUsername.(string)

		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: http.StatusText(http.StatusInternalServerError)})
			return
		}

		// Get the current cookie from the request
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No session cookie found"})
			return
		}

		// Extend the cookie expiration by 24 hours from now
		newExpiration := time.Now().Add(24 * time.Hour)

		// Update the cookie in the database
		loginCookie := database.LoginCookie{
			UserName:   userName,
			Cookie:     cookie.Value,
			Expiration: newExpiration,
		}

		err = loginCookie.Update(r.Context(), apiService.DatabaseClient.DB)
		if err != nil {
			apiService.Error().Msgf("error %s updating session cookie %+v", err, loginCookie)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to refresh session"})
			return
		}

		// Set the new cookie with extended expiration
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    cookie.Value,
			Path:     "/",
			Expires:  newExpiration,
			HttpOnly: true,
			Secure:   false,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Session refreshed successfully",
			"expires": newExpiration.Format(time.RFC3339),
		})
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

		username, ok := r.Context().Value(UsernameContextKey).(string)
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

func CreateLocationsHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)
		username, ok := rawUsername.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: fmt.Sprintf("Locations page for user: %s", username)})
	}
}

func CreateSolarHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)
		username, ok := rawUsername.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: fmt.Sprintf("Solar page for user: %s", username)})
	}
}

func CreateSettingsHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)
		username, ok := rawUsername.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: fmt.Sprintf("Settings page for user: %s", username)})
	}
}

func CreateHomeHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawUsername := r.Context().Value(UsernameContextKey)
		username, ok := rawUsername.(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: fmt.Sprintf("Home page for user: %s", username)})
	}
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

func CreateAddSensorHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			http.Error(w, "failed to get username from context", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var request struct {
			EUI      string `json:"eui"`
			Name     string `json:"name"`
			Location string `json:"location"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to decode request body")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request body"})
			return
		}

		// Validate required fields
		if request.EUI == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "EUI is required"})
			return
		}

		// Set default values if not provided
		if request.Name == "" {
			request.Name = "Sensor " + request.EUI
		}
		if request.Location == "" {
			request.Location = "Unknown Location"
		}

		// Create sensor in database
		sensor, err := database.CreateSensor(ctx, apiService.DatabaseClient.DB, request.EUI, request.Name, request.Location)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to create sensor in database")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to create sensor"})
			return
		}

		log.Ctx(ctx).Info().Str("sensor_id", sensor.ID).Str("username", username).Msg("sensor created successfully")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Sensor created successfully"})
	}
}

func CreateDeleteSensorHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			http.Error(w, "failed to get username from context", http.StatusUnauthorized)
			return
		}

		// Get sensor ID from URL path
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]
		if sensorID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Sensor ID is required"})
			return
		}

		// Check if sensor exists first
		_, err := database.GetSensorByID(ctx, apiService.DatabaseClient.DB, sensorID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("sensor not found: %s", sensorID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Sensor not found"})
			return
		}

		// Delete sensor from database
		err = database.DeleteSensor(ctx, apiService.DatabaseClient.DB, sensorID)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("failed to delete sensor: %s", sensorID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to delete sensor"})
			return
		}

		log.Ctx(ctx).Info().Str("sensor_id", sensorID).Str("username", username).Msg("sensor deleted successfully")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Sensor deleted successfully"})
	}
}

func CreateGetDroneImagesHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse date range from query parameters
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		var start, end time.Time
		var err error

		if startDate != "" {
			start, err = time.Parse(time.RFC3339, startDate)
			if err != nil {
				apiService.Error().Msgf("Invalid start_date format: %s", startDate)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid start_date format. Use RFC3339"})
				return
			}
		}

		if endDate != "" {
			end, err = time.Parse(time.RFC3339, endDate)
			if err != nil {
				apiService.Error().Msgf("Invalid end_date format: %s", endDate)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid end_date format. Use RFC3339"})
				return
			}
		}

		// If no dates provided, use a default range (e.g., last 30 days)
		if startDate == "" && endDate == "" {
			end = time.Now()
			start = end.AddDate(0, -1, 0) // Last 30 days
		}

		dbImages, err := database.GetDroneImagesByDateRange(r.Context(), apiService.DatabaseClient.DB, start, end)
		if err != nil {
			apiService.Error().Msgf("Error retrieving drone images: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to retrieve drone images"})
			return
		}

		response := api.GetDroneImagesResponse{
			Images: make([]api.DroneImage, len(dbImages)),
		}

		// Convert database.DroneImage to api.DroneImage
		for i, img := range dbImages {
			response.Images[i] = api.DroneImage{
				ID:          img.ID.String(),
				FileName:    img.FileName,
				FilePath:    img.FilePath,
				UploadDate:  img.UploadDate,
				FileSize:    img.FileSize,
				MimeType:    img.MimeType,
				Description: img.Description,
				Metadata:    img.Metadata,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateUploadDroneImagesHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form with 32MB max memory
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			apiService.Error().Msgf("Error parsing multipart form: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to parse form data"})
			return
		}

		files := r.MultipartForm.File["images"]
		if len(files) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No images provided"})
			return
		}

		var uploadedImages []api.DroneImage

		// Ensure storage directory exists
		storageDir := "storage/drone_images"
		if err := os.MkdirAll(storageDir, 0755); err != nil {
			apiService.Error().Msgf("Error creating storage directory: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to prepare storage"})
			return
		}

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				apiService.Error().Msgf("Error opening file %s: %s", fileHeader.Filename, err)
				continue
			}
			defer file.Close()

			// Generate unique ID for the image
			imageID := uuid.New()

			// Create storage path
			storagePath := fmt.Sprintf("%s/%s", storageDir, imageID)

			// Create storage file
			dst, err := os.Create(storagePath)
			if err != nil {
				apiService.Error().Msgf("Error creating storage file: %s", err)
				continue
			}
			defer dst.Close()

			// Copy file content to storage
			if _, err := io.Copy(dst, file); err != nil {
				apiService.Error().Msgf("Error copying file content: %s", err)
				os.Remove(storagePath) // Clean up on error
				continue
			}

			// Create database DroneImage record
			dbImage := database.DroneImage{
				ID:          imageID,
				FileName:    fileHeader.Filename,
				FilePath:    storagePath, // Store absolute path
				UploadDate:  time.Now(),
				FileSize:    fileHeader.Size,
				MimeType:    fileHeader.Header.Get("Content-Type"),
				Description: r.FormValue("description"),
				Metadata: map[string]interface{}{
					"original_name": fileHeader.Filename,
				},
			}

			// Parse and merge additional metadata if provided
			if metadataStr := r.FormValue("metadata"); metadataStr != "" {
				var metadata map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
					for k, v := range metadata {
						dbImage.Metadata[k] = v
					}
				}
			}

			// Save image metadata to database
			err = dbImage.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Error saving drone image metadata: %s", err)
				os.Remove(storagePath) // Clean up on error
				continue
			}

			// Convert to API type
			apiImage := api.DroneImage{
				ID:          dbImage.ID.String(),
				FileName:    dbImage.FileName,
				FilePath:    dbImage.FilePath,
				UploadDate:  dbImage.UploadDate,
				FileSize:    dbImage.FileSize,
				MimeType:    dbImage.MimeType,
				Description: dbImage.Description,
				Metadata:    dbImage.Metadata,
			}

			uploadedImages = append(uploadedImages, apiImage)
		}

		if len(uploadedImages) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to upload any images"})
			return
		}

		response := api.UploadDroneImagesResponse{
			UploadedImages: uploadedImages,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateGetDroneImageHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		imageIDStr := vars["image_id"]

		if imageIDStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image ID is required"})
			return
		}

		imageID, err := uuid.Parse(imageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid image ID format"})
			return
		}

		dbImage, err := database.GetDroneImageByID(r.Context(), apiService.DatabaseClient.DB, imageID)
		if err != nil {
			if errors.Is(err, database.ErrorNoDroneImage) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image not found"})
				return
			}
			apiService.Error().Msgf("Error retrieving drone image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to retrieve image"})
			return
		}

		// Convert to API type
		apiImage := api.DroneImage{
			ID:          dbImage.ID.String(),
			FileName:    dbImage.FileName,
			FilePath:    dbImage.FilePath,
			UploadDate:  dbImage.UploadDate,
			FileSize:    dbImage.FileSize,
			MimeType:    dbImage.MimeType,
			Description: dbImage.Description,
			Metadata:    dbImage.Metadata,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(apiImage)
	}
}

func CreateGetDroneImageContentHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		imageIDStr := vars["image_id"]

		if imageIDStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image ID is required"})
			return
		}

		imageID, err := uuid.Parse(imageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid image ID format"})
			return
		}

		// Get image metadata from database
		dbImage, err := database.GetDroneImageByID(r.Context(), apiService.DatabaseClient.DB, imageID)
		if err != nil {
			if errors.Is(err, database.ErrorNoDroneImage) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image not found"})
				return
			}
			apiService.Error().Msgf("Error retrieving drone image metadata: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to retrieve image metadata"})
			return
		}

		// Open and serve the image file
		file, err := os.Open(dbImage.FilePath)
		if err != nil {
			apiService.Error().Msgf("Error opening image file: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to read image file"})
			return
		}
		defer file.Close()

		// Set content type header based on stored mime type
		w.Header().Set("Content-Type", dbImage.MimeType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", dbImage.FileSize))

		// Stream the file content
		if _, err := io.Copy(w, file); err != nil {
			apiService.Error().Msgf("Error streaming image content: %s", err)
			return
		}
	}
}

func CreateGetSensorBatteryDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]

		// Get query parameters for date range
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		var start, end time.Time
		var err error

		if startDate != "" {
			start, err = time.Parse("2006-01-02", startDate)
			if err != nil {
				apiService.Error().Msgf("Invalid start_date format: %s", startDate)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid start_date format. Use YYYY-MM-DD"})
				return
			}
		}

		if endDate != "" {
			end, err = time.Parse("2006-01-02", endDate)
			if err != nil {
				apiService.Error().Msgf("Invalid end_date format: %s", endDate)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid end_date format. Use YYYY-MM-DD"})
				return
			}
			// Set end date to end of day (23:59:59) to include all data for that day
			end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		// Retrieve data for the sensorID
		data, err := database.GetSensorBatteryDataForSensorID(r.Context(), apiService.DatabaseClient.DB, sensorID, start, end)
		if err != nil {
			if errors.Is(err, database.ErrorNoSensorBatteryData) {
				apiService.Debug().Msgf("No battery data found for sensor_id: %s", sensorID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "No data found"})
				return
			}

			apiService.Error().Msgf("Error retrieving battery data for sensor_id: %s, error: %s", sensorID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Convert database data to GetBatteryLevelDataResponse
		response := api.GetBatteryLevelDataResponse{
			BatteryLevelData: make([]api.BatteryLevelData, 0),
		}
		for _, d := range data {
			response.BatteryLevelData = append(response.BatteryLevelData, api.BatteryLevelData{
				Date:         d.Date,
				BatteryLevel: d.BatteryLevel,
			})
		}

		// Send the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func CreateSetSensorBatteryDataHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		sensorID := vars["sensor_id"]

		request := api.SetBatteryLevelDataResponse{}
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

		// Iterate over each BatteryLevelData item and save to the database
		for _, batteryData := range request.BatteryLevelData {
			sensorBatteryData := database.SensorBatteryData{
				SensorID:     sensorID,
				Date:         batteryData.Date,
				BatteryLevel: batteryData.BatteryLevel,
			}

			err = sensorBatteryData.Save(r.Context(), apiService.DatabaseClient.DB)
			if err != nil {
				apiService.Error().Msgf("Failed to save battery data for sensor_id: %s, data: %+v, error: %s", sensorID, sensorBatteryData, err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to save battery data"})
				return
			}
		}

		// Send success response
		apiService.Trace().Msgf("Successfully saved battery data for sensor_id: %s", sensorID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Battery data saved successfully"})
	}
}

func CreateDeleteDroneImageHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		imageIDStr := vars["image_id"]

		if imageIDStr == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image ID is required"})
			return
		}

		imageID, err := uuid.Parse(imageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid image ID format"})
			return
		}

		// Get image metadata to find file path
		dbImage, err := database.GetDroneImageByID(r.Context(), apiService.DatabaseClient.DB, imageID)
		if err != nil {
			if errors.Is(err, database.ErrorNoDroneImage) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image not found"})
				return
			}
			apiService.Error().Msgf("Error retrieving drone image metadata: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to retrieve image metadata"})
			return
		}

		// Delete the image file
		if err := os.Remove(dbImage.FilePath); err != nil && !os.IsNotExist(err) {
			apiService.Error().Msgf("Error deleting image file: %s", err)
			// Continue with database deletion even if file deletion fails
		}

		// Delete from database
		err = database.DeleteDroneImage(r.Context(), apiService.DatabaseClient.DB, imageID)
		if err != nil {
			if errors.Is(err, database.ErrorNoDroneImage) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Image not found"})
				return
			}
			apiService.Error().Msgf("Error deleting drone image: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to delete image"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Image deleted successfully"})
	}
}

// CreateGetAllUsersHandler returns a handler that gets all users (admin only)
func CreateGetAllUsersHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get the current user from the context (set by auth middleware)
		currentUser, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			apiService.Error().Msg("No user found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if current user is admin or root_admin
		userRole, err := database.GetUserRole(ctx, apiService.DatabaseClient.DB, currentUser)
		if err != nil {
			apiService.Error().Msgf("Error getting user role: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		if userRole != "admin" && userRole != "root_admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Insufficient permissions"})
			return
		}

		// Get all users
		users, err := database.GetAllUsers(ctx, apiService.DatabaseClient.DB)
		if err != nil {
			apiService.Error().Msgf("Error getting all users: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to retrieve users"})
			return
		}

		// Convert to API response format
		var userList []api.User
		for _, user := range users {
			userList = append(userList, api.User{
				Username:  user.UserName,
				Role:      user.Role,
				CreatedAt: time.Now().Format(time.RFC3339), // TODO: Add created_at to database
				LastLogin: time.Now().Format(time.RFC3339), // TODO: Add last_login to database
				IsActive:  true,                            // TODO: Add is_active to database
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.ListUsersResponse{Users: userList})
	}
}

// CreateUpdateUserRoleHandler returns a handler that updates a user's role (admin only)
func CreateUpdateUserRoleHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		username := vars["username"]

		// Get the current user from the context
		currentUser, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			apiService.Error().Msg("No user found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if current user is admin or root_admin
		userRole, err := database.GetUserRole(ctx, apiService.DatabaseClient.DB, currentUser)
		if err != nil {
			apiService.Error().Msgf("Error getting user role: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		if userRole != "admin" && userRole != "root_admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Insufficient permissions"})
			return
		}

		// Parse request body
		var request api.UpdateUserRoleRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request body"})
			return
		}

		// Validate role
		if request.Role != "user" && request.Role != "admin" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid role. Must be 'user' or 'admin'"})
			return
		}

		// Check if target user exists
		_, err = database.GetLoginAuthenticationByUserName(ctx, apiService.DatabaseClient.DB, username)
		if err != nil {
			if err == database.ErrorNoLoginAuthenticationForUsername {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "User not found"})
				return
			}
			apiService.Error().Msgf("Error checking if user exists: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Update user role
		err = database.UpdateUserRole(ctx, apiService.DatabaseClient.DB, username, request.Role)
		if err != nil {
			apiService.Error().Msgf("Error updating user role: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to update user role"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "User role updated successfully"})
	}
}

// CreateRemoveAdminHandler returns a handler that removes admin permissions (root_admin only)
func CreateRemoveAdminHandler(apiService *APIService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)
		username := vars["username"]

		// Get the current user from the context
		currentUser, ok := ctx.Value(UsernameContextKey).(string)
		if !ok {
			apiService.Error().Msg("No user found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Unauthorized"})
			return
		}

		// Check if current user is root_admin
		userRole, err := database.GetUserRole(ctx, apiService.DatabaseClient.DB, currentUser)
		if err != nil {
			apiService.Error().Msgf("Error getting user role: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		if userRole != "root_admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Only root admin can remove admin permissions"})
			return
		}

		// Prevent removing admin from root_admin
		if username == currentUser {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Cannot remove admin permissions from root admin"})
			return
		}

		// Check if target user exists and get their current role
		targetUser, err := database.GetLoginAuthenticationByUserName(ctx, apiService.DatabaseClient.DB, username)
		if err != nil {
			if err == database.ErrorNoLoginAuthenticationForUsername {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(api.ErrorResponse{Error: "User not found"})
				return
			}
			apiService.Error().Msgf("Error checking if user exists: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Internal server error"})
			return
		}

		// Only allow removing admin permissions from admin users
		if targetUser.Role != "admin" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "User is not an admin"})
			return
		}

		// Update user role to 'user'
		err = database.UpdateUserRole(ctx, apiService.DatabaseClient.DB, username, "user")
		if err != nil {
			apiService.Error().Msgf("Error removing admin permissions: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Failed to remove admin permissions"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.SuccessResponse{Message: "Admin permissions removed successfully"})
	}
}
