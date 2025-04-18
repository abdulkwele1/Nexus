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

type Sensor struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	InstallationDate  string `json:"installation_date"`
	SensorCoordinates `json:"sensor_coordinates"`
}

type SensorMoistureData struct {
	ID           int       `json:"id"`
	SensorID     int       `json:"sensor_id"`
	Date         time.Time `json:"date"`
	SoilMoisture float64   `json:"soil_moisture"`
}

type SensorTemperatureData struct {
	ID              int       `json:"id"`
	SensorID        int       `json:"sensor_id"`
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
