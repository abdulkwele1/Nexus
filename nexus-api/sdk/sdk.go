// package sdk implements clients for programmatically
// interacting with the REST APIs of the relay services
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

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
	nc.Info().Str("cookie_name", nc.Cookie.Name).Str("cookie_value", nc.Cookie.Value).Msg("Login successful, cookie stored in SDK client")

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
func (nc *NexusClient) GetSensorMoistureData(ctx context.Context, sensorID string) (api.GetSensorMoistureDataResponse, error) {
	endpoint := fmt.Sprintf("%s/sensors/%s/moisture_data", nc.Config.NexusAPIEndpoint, sensorID)

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
func (nc *NexusClient) SetSensorMoistureData(ctx context.Context, sensorID string, moistureData api.SetSensorMoistureDataResponse) error {
	endpoint := fmt.Sprintf("%s/sensors/%s/moisture_data", nc.Config.NexusAPIEndpoint, sensorID)

	body, err := json.Marshal(moistureData)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(httpRequest, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}

// GetSensorTemperatureData retrieves temperature data for a specific sensor
func (nc *NexusClient) GetSensorTemperatureData(ctx context.Context, sensorID string) (api.GetSensorTemperatureDataResponse, error) {
	endpoint := fmt.Sprintf("%s/sensors/%s/temperature_data", nc.Config.NexusAPIEndpoint, sensorID)

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
func (nc *NexusClient) SetSensorTemperatureData(ctx context.Context, sensorID string, temperatureData api.SetSensorTemperatureDataResponse) error {
	endpoint := fmt.Sprintf("%s/sensors/%s/temperature_data", nc.Config.NexusAPIEndpoint, sensorID)

	body, err := json.Marshal(temperatureData)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpRequest.Header.Set("Content-Type", "application/json")
	err = SetAuthHeaders(httpRequest, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
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

// GetAllSensors retrieves all sensors from the API
func (nc *NexusClient) GetAllSensors(ctx context.Context) ([]api.Sensor, error) {
	endpoint := fmt.Sprintf("%s/sensors", nc.Config.NexusAPIEndpoint)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return nil, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return nil, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var sensors []api.Sensor
	err = json.NewDecoder(response.Body).Decode(&sensors)
	if err != nil {
		return nil, err
	}

	return sensors, nil
}

// GetDroneImages retrieves drone images within a date range
func (nc *NexusClient) GetDroneImages(ctx context.Context, startDate, endDate time.Time) (api.GetDroneImagesResponse, error) {
	endpoint := fmt.Sprintf("%s/drone_images?start_date=%s&end_date=%s",
		nc.Config.NexusAPIEndpoint,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339))

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetDroneImagesResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetDroneImagesResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetDroneImagesResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetDroneImagesResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetDroneImagesResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetDroneImagesResponse{}, err
	}

	return result, nil
}

// GetDroneImage retrieves a specific drone image by ID
func (nc *NexusClient) GetDroneImage(ctx context.Context, imageID string) (*api.DroneImage, error) {
	endpoint := fmt.Sprintf("%s/drone_images/%s", nc.Config.NexusAPIEndpoint, imageID)

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return nil, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("image not found")
	}

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return nil, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var image api.DroneImage
	err = json.NewDecoder(response.Body).Decode(&image)
	if err != nil {
		return nil, err
	}

	return &image, nil
}

// UploadDroneImages uploads multiple drone images
func (nc *NexusClient) UploadDroneImages(ctx context.Context, files [][]byte, fileNames []string, description string, metadata map[string]interface{}) (api.UploadDroneImagesResponse, error) {
	endpoint := fmt.Sprintf("%s/drone_images", nc.Config.NexusAPIEndpoint)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add files
	for i, fileData := range files {
		part, err := writer.CreateFormFile("images", fileNames[i])
		if err != nil {
			return api.UploadDroneImagesResponse{}, err
		}
		if _, err := io.Copy(part, bytes.NewReader(fileData)); err != nil {
			return api.UploadDroneImagesResponse{}, err
		}
	}

	// Add description if provided
	if description != "" {
		if err := writer.WriteField("description", description); err != nil {
			return api.UploadDroneImagesResponse{}, err
		}
	}

	// Add metadata if provided
	if metadata != nil {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return api.UploadDroneImagesResponse{}, err
		}
		if err := writer.WriteField("metadata", string(metadataJSON)); err != nil {
			return api.UploadDroneImagesResponse{}, err
		}
	}

	if err := writer.Close(); err != nil {
		return api.UploadDroneImagesResponse{}, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, body)
	if err != nil {
		return api.UploadDroneImagesResponse{}, err
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())
	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.UploadDroneImagesResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.UploadDroneImagesResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.UploadDroneImagesResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.UploadDroneImagesResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.UploadDroneImagesResponse{}, err
	}

	return result, nil
}

// DeleteDroneImage deletes a specific drone image by ID
func (nc *NexusClient) DeleteDroneImage(ctx context.Context, imageID string) error {
	endpoint := fmt.Sprintf("%s/drone_images/%s", nc.Config.NexusAPIEndpoint, imageID)

	request, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("image not found")
	}

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	return nil
}

// SetAuthHeaders sets the headers needed to authenticate requests
// to the Nexus API, returning error (if any)
func SetAuthHeaders(request *http.Request, cookie *http.Cookie) error {
	if cookie == nil || cookie.Value == "" {
		return fmt.Errorf("invalid auth cookie (nil or empty)")
	}
	request.AddCookie(cookie)
	fmt.Printf("[TRACE] SetAuthHeaders adding cookie: Name=%s, Value=%s\n", cookie.Name, cookie.Value)
	return nil
}

// NewClient creates a new client using the provided configuration
// returning the client and error (if any)
// GetSensorBatteryData retrieves battery data for a specific sensor
func (nc *NexusClient) GetSensorBatteryData(ctx context.Context, sensorID string, startDate, endDate string) (api.GetBatteryLevelDataResponse, error) {
	endpoint := fmt.Sprintf("%s/sensors/%s/battery", nc.Config.NexusAPIEndpoint, sensorID)

	// Add query parameters for date range if provided
	if startDate != "" || endDate != "" {
		params := make([]string, 0, 2)
		if startDate != "" {
			params = append(params, fmt.Sprintf("start_date=%s", startDate))
		}
		if endDate != "" {
			params = append(params, fmt.Sprintf("end_date=%s", endDate))
		}
		if len(params) > 0 {
			endpoint = fmt.Sprintf("%s?%s", endpoint, strings.Join(params, "&"))
		}
	}

	request, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return api.GetBatteryLevelDataResponse{}, err
	}

	err = SetAuthHeaders(request, nc.Cookie)
	if err != nil {
		return api.GetBatteryLevelDataResponse{}, err
	}

	response, err := nc.http.Do(request)
	if err != nil {
		return api.GetBatteryLevelDataResponse{}, err
	}
	defer response.Body.Close()

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.GetBatteryLevelDataResponse{}, fmt.Errorf("non 200-level status code: %d", response.StatusCode)
	}

	var result api.GetBatteryLevelDataResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return api.GetBatteryLevelDataResponse{}, err
	}

	return result, nil
}

// SetSensorBatteryData saves battery data for a specific sensor
func (nc *NexusClient) SetSensorBatteryData(ctx context.Context, sensorID string, batteryData api.SetBatteryLevelDataResponse) error {
	endpoint := fmt.Sprintf("%s/sensors/%s/battery", nc.Config.NexusAPIEndpoint, sensorID)

	body, err := json.Marshal(batteryData)
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

func NewClient(config SDKConfig) (*NexusClient, error) {
	client := NexusClient{
		http:          http.Client{},
		Config:        config,
		ServiceLogger: config.Logger,
		Cookie:        &http.Cookie{},
	}

	return &client, nil
}
