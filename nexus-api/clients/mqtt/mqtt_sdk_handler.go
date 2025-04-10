package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nexus-api/api"
	"nexus-api/logging"
	"nexus-api/sdk"
)

// SensorMessage represents the structure of incoming sensor data messages
type SensorMessage struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

// SensorType represents the different types of sensors
const (
	SensorTypeMoisture    = "4102" // Soil moisture sensor
	SensorTypeTemperature = "4103" // Soil temperature sensor
)

// SDKMQTTHandler wraps an SDK client for MQTT message handling
type SDKMQTTHandler struct {
	sdkClient *sdk.NexusClient
	logger    *logging.ServiceLogger
}

// NewSDKMQTTHandler creates a new MQTT handler that uses the SDK client
func NewSDKMQTTHandler(sdkClient *sdk.NexusClient, logger *logging.ServiceLogger) *SDKMQTTHandler {
	return &SDKMQTTHandler{
		sdkClient: sdkClient,
		logger:    logger,
	}
}

// getSensorName returns a formatted sensor name based on the sensor ID
func getSensorName(sensorID int) string {
	return fmt.Sprintf("Sensor%d", sensorID)
}

// HandleSensorData processes sensor data messages and forwards them to the SDK
func (h *SDKMQTTHandler) HandleSensorData(ctx context.Context, topic string, payload []byte) error {
	// Parse the sensor message from the payload
	var sensorMsg SensorMessage
	if err := json.Unmarshal(payload, &sensorMsg); err != nil {
		return fmt.Errorf("failed to parse sensor message: %w", err)
	}

	// Extract sensor ID and type from topic
	// Format: /device_sensor_data/{deviceID}/{sensorID}/{type}/{subtype}/{sensorType}
	parts := strings.Split(topic, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	// The sensor ID is the third value (index 2)
	sensorID := parts[2]

	// The sensor type is the last value
	sensorType := parts[len(parts)-1]

	// Convert sensor ID to integer if needed
	sensorIDInt, err := strconv.Atoi(sensorID)
	if err != nil {
		h.logger.Warn().Err(err).Msgf("Sensor ID %s is not a number, using default ID", sensorID)
		sensorIDInt = 0
	}

	// Generate sensor name using the new naming convention
	sensorName := getSensorName(sensorIDInt)
	h.logger.Info().Msgf("Processing sensor data - Name: %s, Type: %s", sensorName, sensorType)

	// Convert timestamp from milliseconds to time.Time
	timestamp := time.UnixMilli(sensorMsg.Timestamp)

	// Create sensor data based on the sensor type
	switch sensorType {
	case SensorTypeMoisture:
		moistureData := api.SetSensorMoistureDataResponse{
			SensorMoistureData: []api.SensorMoistureData{
				{
					SensorID:     sensorIDInt,
					Date:         timestamp.Format(time.RFC3339),
					SoilMoisture: sensorMsg.Value,
				},
			},
		}
		if err := h.sdkClient.SetSensorMoistureData(ctx, sensorIDInt, moistureData); err != nil {
			return fmt.Errorf("failed to save moisture data: %w", err)
		}
		h.logger.Info().Msgf("Saved moisture data for %s", sensorName)

	case SensorTypeTemperature:
		temperatureData := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{
				{
					SensorID:        sensorIDInt,
					Date:            timestamp.Format(time.RFC3339),
					SoilTemperature: sensorMsg.Value,
				},
			},
		}
		if err := h.sdkClient.SetSensorTemperatureData(ctx, sensorIDInt, temperatureData); err != nil {
			return fmt.Errorf("failed to save temperature data: %w", err)
		}
		h.logger.Info().Msgf("Saved temperature data for %s", sensorName)

	default:
		h.logger.Warn().Msgf("Unknown sensor type: %s, skipping data", sensorType)
	}

	return nil
}

// CreateSDKMessageHandler creates an MQTT message handler that uses the SDK client
func CreateSDKMessageHandler(ctx context.Context, sdkClient *sdk.NexusClient, logger *logging.ServiceLogger) MQTTMessageHandler {
	handler := NewSDKMQTTHandler(sdkClient, logger)

	return func(ctx context.Context, topic string, payload []byte, logger *logging.ServiceLogger) error {
		// Handle different topics
		switch {
		case strings.HasPrefix(topic, "/device_sensor_data/"):
			return handler.HandleSensorData(ctx, topic, payload)
		default:
			return fmt.Errorf("unhandled topic: %s", topic)
		}
	}
}
