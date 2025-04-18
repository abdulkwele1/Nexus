package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nexus-api/logging"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTMessageHandler is a function that processes MQTT messages
type MQTTMessageHandler func(ctx context.Context, topic string, payload []byte, logger *logging.ServiceLogger) error

// CreateMQTTMessageHandler creates a message handler for the MQTT client
func CreateMQTTMessageHandler(ctx context.Context, logger *logging.ServiceLogger, handlers map[string]MQTTMessageHandler, overrideHandlers map[string]MQTTMessageHandler) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		payload := msg.Payload()

		logger.Info().Msgf("Received message on topic %s: %s", topic, string(payload))

		// Find the appropriate handler for this topic
		var handler MQTTMessageHandler

		// Check for override handler first
		if overrideHandlers != nil {
			if h, exists := overrideHandlers[topic]; exists {
				handler = h
			}
		}

		// If no override handler, check for exact match
		if handler == nil {
			if h, exists := handlers[topic]; exists {
				handler = h
			}
		}

		// If still no handler, check for wildcard match
		if handler == nil {
			for pattern, h := range handlers {
				if strings.Contains(pattern, "+") || strings.Contains(pattern, "#") {
					if matchTopic(pattern, topic) {
						handler = h
						break
					}
				}
			}
		}

		if handler == nil {
			logger.Warn().Msgf("No handler found for topic: %s", topic)
			return
		}

		// Process the message
		err := handler(ctx, topic, payload, logger)
		if err != nil {
			logger.Error().Err(err).Msgf("Error processing message on topic %s", topic)
		} else {
			logger.Info().Msgf("Successfully processed message on topic %s", topic)
		}
	}
}

// matchTopic checks if a topic matches a pattern with wildcards
func matchTopic(pattern, topic string) bool {
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	if len(patternParts) != len(topicParts) && !strings.Contains(pattern, "#") {
		return false
	}

	for i := 0; i < len(patternParts) && i < len(topicParts); i++ {
		if patternParts[i] == "#" {
			return true
		}
		if patternParts[i] == "+" {
			continue
		}
		if patternParts[i] != topicParts[i] {
			return false
		}
	}

	return true
}

// DefaultSensorDataHandler is a default handler for sensor data messages
func DefaultSensorDataHandler(ctx context.Context, topic string, payload []byte, logger *logging.ServiceLogger) error {
	// Extract sensor ID from topic (assuming format: sensors/{id}/data)
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	sensorID := parts[1]

	// Parse the payload
	var data map[string]interface{}
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return fmt.Errorf("failed to parse sensor data: %w", err)
	}

	// Add timestamp if not present
	if _, exists := data["timestamp"]; !exists {
		data["timestamp"] = time.Now().Format(time.RFC3339)
	}

	// Log the sensor data
	logger.Info().Interface("data", data).Msgf("Received data from sensor %s", sensorID)

	// Here you would typically process the sensor data
	// For example, store it in a database, trigger alerts, etc.

	return nil
}

// DefaultSystemStatusHandler is a default handler for system status messages
func DefaultSystemStatusHandler(ctx context.Context, topic string, payload []byte, logger *logging.ServiceLogger) error {
	// Parse the payload
	var status map[string]interface{}
	err := json.Unmarshal(payload, &status)
	if err != nil {
		return fmt.Errorf("failed to parse system status: %w", err)
	}

	// Log the system status
	logger.Info().Interface("status", status).Msg("Received system status update")

	// Here you would typically process the system status
	// For example, update internal state, trigger alerts, etc.

	return nil
}
