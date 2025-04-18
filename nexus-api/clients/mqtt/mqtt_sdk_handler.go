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

	h.logger.Info().Msgf("Processing sensor data - ID: %s, Type: %s", sensorID, sensorType)

	// Convert sensor ID to integer if needed
	// Note: If your sensor IDs are not numeric, you might need to handle them differently
	sensorIDInt, err := strconv.Atoi(sensorID)
	if err != nil {
		// If the sensor ID is not a number, we'll use a hash or other method to generate a numeric ID
		// For now, we'll just log the error and use a default value
		h.logger.Warn().Err(err).Msgf("Sensor ID %s is not a number, using default ID", sensorID)
		sensorIDInt = 0 // Default ID, you might want to use a different approach
	}

	// Convert timestamp from milliseconds to time.Time
	timestamp := time.UnixMilli(sensorMsg.Timestamp)

	// Create sensor data based on the sensor type
	switch sensorType {
	case "4102": // Moisture sensor
		moistureData := api.SetSensorMoistureDataResponse{
			SensorMoistureData: []api.SensorMoistureData{
				{
					Date:         timestamp,
					SoilMoisture: sensorMsg.Value,
				},
			},
		}
		if err := h.sdkClient.SetSensorMoistureData(ctx, sensorIDInt, moistureData); err != nil {
			return fmt.Errorf("failed to save moisture data: %w", err)
		}
		h.logger.Info().Msgf("Saved moisture data for sensor %s", sensorID)
	case "4103": // Temperature sensor
		temperatureData := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{
				{
					Date:            timestamp,
					SoilTemperature: sensorMsg.Value,
				},
			},
		}
		if err := h.sdkClient.SetSensorTemperatureData(ctx, sensorIDInt, temperatureData); err != nil {
			return fmt.Errorf("failed to save temperature data: %w", err)
		}
		h.logger.Info().Msgf("Saved temperature data for sensor %s", sensorID)
	default:
		return fmt.Errorf("unknown sensor type: %s", sensorType)
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
