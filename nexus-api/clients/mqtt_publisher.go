package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nexus-api/logging"
)

// MQTTMessage represents a message to be published to an MQTT topic
type MQTTMessage struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
	QoS     byte        `json:"qos"`
	Retain  bool        `json:"retain"`
}

// PublishToMQTT publishes a message to an MQTT topic
func PublishToMQTT(ctx context.Context, client *MQTTClient, message MQTTMessage) error {
	if client == nil {
		return fmt.Errorf("MQTT client is nil")
	}

	if !client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	// Convert payload to JSON if it's not already a string or byte slice
	var payload interface{}
	switch v := message.Payload.(type) {
	case string:
		payload = v
	case []byte:
		payload = v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal payload to JSON: %w", err)
		}
		payload = jsonBytes
	}

	// Set default QoS if not provided
	qos := message.QoS
	if qos == 0 {
		qos = 1
	}

	// Publish the message
	err := client.Publish(ctx, message.Topic, qos, message.Retain, payload)
	if err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", message.Topic, err)
	}

	return nil
}

// PublishSensorData publishes sensor data to an MQTT topic
func PublishSensorData(ctx context.Context, client *MQTTClient, sensorID string, data interface{}, logger *logging.ServiceLogger) error {
	if client == nil {
		return fmt.Errorf("MQTT client is nil")
	}

	topic := fmt.Sprintf("sensors/%s/data", sensorID)
	message := MQTTMessage{
		Topic:   topic,
		Payload: data,
		QoS:     1,
		Retain:  false,
	}

	err := PublishToMQTT(ctx, client, message)
	if err != nil {
		logger.Error().Err(err).Msgf("Failed to publish sensor data for sensor %s", sensorID)
		return err
	}

	logger.Info().Msgf("Published sensor data for sensor %s to topic %s", sensorID, topic)
	return nil
}

// PublishSystemStatus publishes system status to an MQTT topic
func PublishSystemStatus(ctx context.Context, client *MQTTClient, status map[string]interface{}, logger *logging.ServiceLogger) error {
	if client == nil {
		return fmt.Errorf("MQTT client is nil")
	}

	// Add timestamp to status
	status["timestamp"] = time.Now().Format(time.RFC3339)

	message := MQTTMessage{
		Topic:   "system/status",
		Payload: status,
		QoS:     1,
		Retain:  true, // Retain the last status message
	}

	err := PublishToMQTT(ctx, client, message)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to publish system status")
		return err
	}

	logger.Info().Msg("Published system status")
	return nil
}
