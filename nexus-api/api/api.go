package api

import (
	"time"
)

type UserCookies = map[string]string

type LoginRequest struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginResponse struct {
	RedirectURL string `json:"redirect_url"`
	Match       bool   `json:"match"`
	Cookie      string `json:"cookie"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type Panel struct {
	PanelID          int     `json:"panel_id"`          // Unique ID of the panel
	Name             string  `json:"name"`              // Name of the panel
	Location         string  `json:"location"`          // Location of the panel
	InstallationDate string  `json:"installation_date"` // Installation date in YYYY-MM-DD format
	CapacityKW       float64 `json:"capacity_kW"`       // Capacity in kilowatts
}

// UpdatePanelRequest represents the payload for updating panel details.
type UpdatePanelRequest struct {
	Name             string  `json:"name"`              // Name of the panel
	Location         string  `json:"location"`          // Location of the panel
	InstallationDate string  `json:"installation_date"` // Installation date in YYYY-MM-DD format
	CapacityKW       float64 `json:"capacity_kW"`       // Capacity in kilowatts

}

// ErrorResponse represents a common error response structure.
type ErrorResponse struct {
	Error string `json:"error"` // Error message
}

// SuccessResponse represents a generic success response structure.
type SuccessResponse struct {
	Message string `json:"message"` // Success message
}

type YieldData struct {
	Date     time.Time `json:"date"`
	KwhYield float64   `json:"kwh_yield"`
}

type GetPanelYieldDataResponse struct {
	YieldData []YieldData `json:"yield_data"`
}

type SetPanelYieldDataResponse struct {
	YieldData []YieldData `json:"yield_data"`
}

type GetSensorMoistureDataResponse struct {
	SensorMoistureData []SensorMoistureData `json:"sensor_moisture_data"`
}

type GetSensorTemperatureDataResponse struct {
	SensorTemperatureData []SensorTemperatureData `json:"sensor_temperature_data"`
}
type ConsumptionData struct {
	Date        time.Time `json:"date"`
	CapacityKwh float64   `json:"capacity_kwh"`
	ConsumedKwh float64   `json:"consumed_kwh"`
}

type GetPanelConsumptionDataResponse struct {
	ConsumptionData []ConsumptionData `json:"consumption_data"`
}

type SetPanelConsumptionDataResponse struct {
	ConsumptionData []ConsumptionData `json:"consumption_data"`
}

type GetAllSensorsResponse struct {
	Sensors []Sensor `json:"sensors"`
}

type Sensor struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Location         string    `json:"location"`
	InstallationDate time.Time `json:"installation_date"`
	SensorCoordinates
}

type SensorMoistureData struct {
	ID           int       `json:"id"`
	SensorID     string    `json:"sensor_id"`
	Date         time.Time `json:"date"`
	SoilMoisture float64   `json:"soil_moisture"`
}

type SensorReading struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"` // Assuming the timestamp is Unix milliseconds
}

type SensorTemperatureData struct {
	ID              int       `json:"id"`
	SensorID        string    `json:"sensor_id"`
	Date            time.Time `json:"date"`
	SoilTemperature float64   `json:"soil_temperature"`
}

type SetSensorMoistureDataResponse struct {
	SensorMoistureData []SensorMoistureData `json:"sensor_moisture_data"`
}

type SetSensorTemperatureDataResponse struct {
	SensorTemperatureData []SensorTemperatureData `json:"sensor_temperature_data"`
}

type SensorCoordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// DroneImage represents a drone image record
type DroneImage struct {
	ID          string                 `json:"id"`
	FileName    string                 `json:"file_name"`
	FilePath    string                 `json:"file_path"`
	UploadDate  time.Time              `json:"upload_date"`
	FileSize    int64                  `json:"file_size"`
	MimeType    string                 `json:"mime_type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type GetDroneImagesResponse struct {
	Images []DroneImage `json:"images"`
}

type UploadDroneImagesResponse struct {
	UploadedImages []DroneImage `json:"uploaded_images"`
}
