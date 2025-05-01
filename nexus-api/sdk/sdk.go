// package sdk implements clients for programmatically
// interacting with the REST APIs of the relay services
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"nexus-api/api"
	"nexus-api/logging"
)

const (
	HealthCheckPath    = "/healthcheck"
	LoginPath          = "/login"
	ChangePasswordPath = "/change-password"
	LogoutPath         = "/logout"
)

// SDKConfig wraps values for configuring
// a NexusDK client
type SDKConfig struct {
	NexusAPIEndpoint string
	UserName         string
	Password         string
	Logger           *logging.ServiceLogger
}

// NexusClient is a client for making requests to
// the Nexus API
type NexusClient struct {
	http   http.Client
	Config SDKConfig
	*logging.ServiceLogger
	Cookie *http.Cookie
}

// ReceiverServiceHealthCheck calls the service check endpoint
// for the Nexus API, returning error for any response with
// a non-200 status code
func (nc *NexusClient) HealthCheck(ctx context.Context) error {
	request, err := http.NewRequest("GET", nc.Config.NexusAPIEndpoint+HealthCheckPath, nil)

	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}

// Login attempts to login to nexus app using username and password
// returning error (if any)
func (nc *NexusClient) Login(ctx context.Context, params api.LoginRequest) (api.LoginResponse, error) {
	body, err := json.Marshal(&params)

	if err != nil {
		return api.LoginResponse{}, err
	}

	request, err := http.NewRequest("POST", nc.Config.NexusAPIEndpoint+LoginPath, bytes.NewBuffer(body))

	if err != nil {
		return api.LoginResponse{}, err
	}

	nc.Trace().Msgf("sending request with params %+v\n headers %+v", params, request.Header)
	response, err := nc.http.Do(request)

	if err != nil {
		return api.LoginResponse{}, err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		// Read and log the response body for more details on the error
		bodyBytes, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			nc.Warn().Err(readErr).Int("status_code", response.StatusCode).Msg("Failed to read response body on error during login")
		} else {
			nc.Error().Int("status_code", response.StatusCode).Str("response_body", string(bodyBytes)).Msg("API returned non-2xx status during login")
		}
		return api.LoginResponse{}, fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	var result api.LoginResponse
	err = json.NewDecoder(response.Body).Decode(&result)

	if err != nil {
		return api.LoginResponse{}, err
	}
	// save cookie for client
	nc.Cookie.Value = result.Cookie
	nc.Cookie.Name = "session_id"

	return result, nil
}

// ChangePassword attempts to change the users password
// returning error (if any)
func (nc *NexusClient) ChangePassword(ctx context.Context, params api.ChangePasswordRequest) error {
	body, err := json.Marshal(&params)

	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", nc.Config.NexusAPIEndpoint+ChangePasswordPath, bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	err = SetAuthHeaders(request, nc.Cookie)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("sending request with params %+v\n headers %+v", params, request.Header)
	response, err := nc.http.Do(request)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}

// Logout attempts to log the user out of the app
// returning error (if any)
func (nc *NexusClient) Logout(ctx context.Context) error {
	request, err := http.NewRequest("GET", nc.Config.NexusAPIEndpoint+LogoutPath, nil)

	if err != nil {
		return err
	}

	err = SetAuthHeaders(request, nc.Cookie)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("sending request %+v\n headers %+v", request, request.Header)
	response, err := nc.http.Do(request)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}

// GetPanelYieldData retrieves yield data for a specific panel
func (nc *NexusClient) GetPanelYieldData(ctx context.Context, panelID int) (api.GetPanelYieldDataResponse, error) {
	endpoint := fmt.Sprintf("%s/panels/%d/yield_data", nc.Config.NexusAPIEndpoint, panelID)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetPanelYieldDataResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetPanelYieldDataResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetPanelYieldDataResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetPanelYieldDataResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetPanelYieldDataResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetPanelYieldDataResponse{}, err
	}

	return result, nil
}

// SetPanelYieldData saves yield data for a specific panel
func (nc *NexusClient) SetPanelYieldData(ctx context.Context, panelID int, yieldData api.SetPanelYieldDataResponse) error {
	endpoint := fmt.Sprintf("%s/panels/%d/yield_data", nc.Config.NexusAPIEndpoint, panelID)

	body, err := json.Marshal(yieldData)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	return nil
}

// GetSensorMoistureData retrieves moisture data for a specific sensor
func (nc *NexusClient) GetSensorMoistureData(ctx context.Context, sensorID int) (api.GetSensorMoistureDataResponse, error) {
	endpoint := fmt.Sprintf("%s/sensors/%d/moisture_data", nc.Config.NexusAPIEndpoint, sensorID)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetSensorMoistureDataResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetSensorMoistureDataResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetSensorMoistureDataResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetSensorMoistureDataResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetSensorMoistureDataResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetSensorMoistureDataResponse{}, err
	}

	return result, nil
}

// SetSensorMoistureData saves moisture data for a specific sensor
func (nc *NexusClient) SetSensorMoistureData(ctx context.Context, sensorID int, moistureData api.SetSensorMoistureDataResponse) error {
	endpoint := fmt.Sprintf("%s/sensors/%d/moisture_data", nc.Config.NexusAPIEndpoint, sensorID)

	body, err := json.Marshal(moistureData)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	return nil
}

// GetSensorTemperatureData retrieves temperature data for a specific sensor
func (nc *NexusClient) GetSensorTemperatureData(ctx context.Context, sensorID int) (api.GetSensorTemperatureDataResponse, error) {
	endpoint := fmt.Sprintf("%s/sensors/%d/temperature_data", nc.Config.NexusAPIEndpoint, sensorID)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetSensorTemperatureDataResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetSensorTemperatureDataResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetSensorTemperatureDataResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetSensorTemperatureDataResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetSensorTemperatureDataResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetSensorTemperatureDataResponse{}, err
	}

	return result, nil
}

// SetSensorTemperatureData saves temperature data for a specific sensor
func (nc *NexusClient) SetSensorTemperatureData(ctx context.Context, sensorID int, temperatureData api.SetSensorTemperatureDataResponse) error {
	endpoint := fmt.Sprintf("%s/sensors/%d/temperature_data", nc.Config.NexusAPIEndpoint, sensorID)

	body, err := json.Marshal(temperatureData)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	return nil
}

// GetPanelConsumptionData retrieves consumption data for a specific panel
func (nc *NexusClient) GetPanelConsumptionData(ctx context.Context, panelID int) (api.GetPanelConsumptionDataResponse, error) {
	endpoint := fmt.Sprintf("%s/panels/%d/consumption_data", nc.Config.NexusAPIEndpoint, panelID)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetPanelConsumptionDataResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetPanelConsumptionDataResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetPanelConsumptionDataResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetPanelConsumptionDataResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetPanelConsumptionDataResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetPanelConsumptionDataResponse{}, err
	}

	return result, nil
}

// SetPanelConsumptionData saves Consumption data for a specific panel
func (nc *NexusClient) SetPanelConsumptionData(ctx context.Context, panelID int, ConsumptionData api.SetPanelConsumptionDataResponse) error {
	endpoint := fmt.Sprintf("%s/panels/%d/consumption_data", nc.Config.NexusAPIEndpoint, panelID)

	body, err := json.Marshal(ConsumptionData)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	return nil
}

// SetAuthHeaders sets the headers needed to authenticate requests
// to the Nexus API, returning error (if any)
func SetAuthHeaders(request *http.Request, cookie *http.Cookie) error {
	request.AddCookie(cookie)
	return nil
}

// NewClient creates a new client using the provided configuration
// returning the client and error (if any)
func NewClient(config SDKConfig) (*NexusClient, error) {
	client := NexusClient{
		http:          http.Client{},
		Config:        config,
		ServiceLogger: config.Logger,
		Cookie:        &http.Cookie{},
	}

	config.Logger.Info().Msgf("Initialized SDK client config - Endpoint: %s, Username: %s", config.NexusAPIEndpoint, config.UserName) // Print config for debugging

	return &client, nil
}
