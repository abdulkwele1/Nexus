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

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConfig contains values for creating a new connection to an MQTT broker
type MQTTConfig struct {
	BrokerURL      string
	ClientID       string
	Username       string
	Password       string
	CleanSession   bool
	AutoReconnect  bool
	ConnectTimeout time.Duration
	Logger         *logging.ServiceLogger
	SDKClient      *sdk.NexusClient
}

// MQTTClient wraps a connection to an MQTT broker
type MQTTClient struct {
	client    mqtt.Client
	logger    *logging.ServiceLogger
	sdkClient *sdk.NexusClient
}

// NewMQTTClient returns a new connection to the specified MQTT broker and error (if any)
func NewMQTTClient(config MQTTConfig) (*MQTTClient, error) {
	// Set default values if not provided
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}

	// Configure MQTT client options
	opts := mqtt.NewClientOptions().
		AddBroker(config.BrokerURL).
		SetClientID(config.ClientID).
		SetCleanSession(config.CleanSession).
		SetAutoReconnect(config.AutoReconnect).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			config.Logger.Error().Err(err).Msg("MQTT connection lost")
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			config.Logger.Info().Msg("MQTT connected")
		})

	// Set credentials if provided
	if config.Username != "" && config.Password != "" {
		opts.SetUsername(config.Username)
		opts.SetPassword(config.Password)
	}

	// Create the client
	client := mqtt.NewClient(opts)

	// Connect to the broker
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return &MQTTClient{
		client:    client,
		logger:    config.Logger,
		sdkClient: config.SDKClient,
	}, nil
}

// Subscribe subscribes to the specified topic with the given QoS level
// and message handler
func (m *MQTTClient) Subscribe(ctx context.Context, topic string, qos byte, callback mqtt.MessageHandler) error {
	token := m.client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}

	m.logger.Info().Msgf("Subscribed to topic: %s", topic)
	return nil
}

// Unsubscribe unsubscribes from the specified topic
func (m *MQTTClient) Unsubscribe(ctx context.Context, topic string) error {
	token := m.client.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, token.Error())
	}

	m.logger.Info().Msgf("Unsubscribed from topic: %s", topic)
	return nil
}

// Publish publishes a message to the specified topic with the given QoS level
func (m *MQTTClient) Publish(ctx context.Context, topic string, qos byte, retained bool, payload interface{}) error {
	token := m.client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, token.Error())
	}

	m.logger.Info().Msgf("Published to topic: %s", topic)
	return nil
}

// Disconnect disconnects from the MQTT broker
func (m *MQTTClient) Disconnect() {
	if m.client.IsConnected() {
		m.client.Disconnect(250)
		m.logger.Info().Msg("Disconnected from MQTT broker")
	}
}

// IsConnected returns true if the client is connected to the MQTT broker
func (m *MQTTClient) IsConnected() bool {
	return m.client.IsConnected()
}

// HealthCheck returns an error if the client is not connected to the MQTT broker
func (m *MQTTClient) HealthCheck() error {
	if !m.client.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}
	return nil
}

// Simple struct to match the expected MQTT JSON payload
type mqttPayload struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"` // Assuming timestamp is epoch milliseconds
}

// HandleMessage implements mqtt.MessageHandler, determines data type from topic,
// parses payload, and calls the appropriate SDK method.
func (m *MQTTClient) HandleMessage(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payloadBytes := msg.Payload()
	ctx := context.Background() // Consider deriving context if needed

	m.logger.Info().Str("topic", topic).Str("payload", string(payloadBytes)).Msg("Received MQTT message")

	// Expected topic structure: /device_sensor_data/{deviceID}/{sensorID_numeric}/{sensorID_hex?}/{channel?}/{typeCode}/{value?}
	// We primarily use parts[2] (numeric ID) and parts[5] (typeCode)
	parts := strings.Split(topic, "/")

	// Basic format check
	if len(parts) < 6 || parts[0] != "" || parts[1] != "device_sensor_data" {
		m.logger.Warn().Str("topic", topic).Msg("Received message on unexpected topic format")
		return
	}

	sensorIDStr := parts[2]
	dataTypeCode := parts[5]

	// Convert sensor ID string to integer
	sensorIDInt, err := strconv.Atoi(sensorIDStr)
	if err != nil {
		m.logger.Error().Err(err).Str("topic", topic).Str("sensorIDStr", sensorIDStr).Msg("Failed to convert sensor ID to integer")
		return
	}

	// Unmarshal the payload JSON
	var receivedData mqttPayload
	err = json.Unmarshal(payloadBytes, &receivedData)
	if err != nil {
		m.logger.Error().Err(err).Str("topic", topic).Str("payload", string(payloadBytes)).Msg("Failed to unmarshal payload JSON")
		return
	}

	// Use the value extracted from JSON
	sensorValue := receivedData.Value
	currentTime := time.Now() // Or potentially use receivedData.Timestamp if needed and converted

	// Process based on the type code from the topic
	switch dataTypeCode {
	case "vt": // Type code for temperature
		data := api.SensorTemperatureData{
			SensorID:        sensorIDInt,
			Date:            currentTime,
			SoilTemperature: sensorValue,
		}
		requestData := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{data},
		}

		err = m.sdkClient.SetSensorTemperatureData(ctx, sensorIDInt, requestData)
		if err != nil {
			m.logger.Error().Err(err).Int("sensorID", sensorIDInt).Msg("Failed to set sensor temperature data via SDK")
			return
		}
		m.logger.Info().Int("sensorID", sensorIDInt).Interface("data", requestData).Msg("Successfully processed and sent temperature data")

	case "vs": // Type code for moisture
		data := api.SensorMoistureData{
			SensorID:     sensorIDInt,
			Date:         currentTime,
			SoilMoisture: sensorValue,
		}
		requestData := api.SetSensorMoistureDataResponse{
			SensorMoistureData: []api.SensorMoistureData{data},
		}

		err = m.sdkClient.SetSensorMoistureData(ctx, sensorIDInt, requestData)
		if err != nil {
			m.logger.Error().Err(err).Int("sensorID", sensorIDInt).Msg("Failed to set sensor moisture data via SDK")
			return
		}
		m.logger.Info().Int("sensorID", sensorIDInt).Interface("data", requestData).Msg("Successfully processed and sent moisture data")

	default:
		m.logger.Warn().Str("topic", topic).Str("typeCode", dataTypeCode).Msg("Received message with unhandled data type code")
	}
}
