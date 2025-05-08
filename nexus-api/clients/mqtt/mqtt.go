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
	BrokerURL string

	ClientID string

	Username string

	Password string

	CleanSession bool

	AutoReconnect bool

	ConnectTimeout time.Duration

	Logger *logging.ServiceLogger

	SDKClient *sdk.NexusClient
}

// MQTTClient wraps a connection to an MQTT broker

type MQTTClient struct {
	client mqtt.Client

	logger *logging.ServiceLogger

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

		client: client,

		logger: config.Logger,

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

// HandleMessage implements mqtt.MessageHandler, determines data type from topic,

// parses payload, and calls the appropriate SDK method.

func (m *MQTTClient) HandleMessage(client mqtt.Client, msg mqtt.Message) {

	topic := msg.Topic()

	payload := msg.Payload()

	ctx := context.Background() // Or derive context from elsewhere if appropriate

	m.logger.Info().Str("topic", topic).Msg("Received message")

	m.logger.Trace().Str("topic", topic).Str("payload", string(payload)).Msg("Raw MQTT Payload")

	// New expected topic structure: /device_sensor_data/{deviceID}/{sensorID_numeric}/{sensorID_hex?}/{channel?}/{typeCode}/{value?}

	// Example: /device_sensor_data/444574498032128/2CF7F1C0627000B2/1/vs/4103

	parts := strings.Split(topic, "/")

	// Check format: needs at least 7 parts (due to leading / and new value identifier) and specific prefix
	if len(parts) < 7 || parts[0] != "" || parts[1] != "device_sensor_data" {
		m.logger.Warn().Str("topic", topic).Msg("Received message on unexpected topic format or insufficient parts")
		return
	}

	// Assuming the numeric ID (parts[2]) is the one needed by the SDK
	sensorIDStr := parts[2]
	// The value identifier (parts[6]) determines the data type (e.g., 4102 for temp, 4103 for moisture)
	valueIdentifier := parts[6]

	// Convert sensor ID string to integer
	sensorIDInt, err := strconv.Atoi(sensorIDStr)

	if err != nil {

		m.logger.Error().Err(err).Str("topic", topic).Str("sensorIDStr", sensorIDStr).Msg("Failed to convert sensor ID to integer")

		return

	}

	// Process based on the value identifier from the topic
	switch valueIdentifier {
	case "4102": // Code for temperature
		var reading api.SensorReading // Use SensorReading for the raw payload
		err := json.Unmarshal(payload, &reading)
		if err != nil {
			m.logger.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal temperature sensor reading")
			return
		}

		ts := time.UnixMilli(reading.Timestamp) // Convert timestamp

		// Construct the SensorTemperatureData object
		tempDetail := api.SensorTemperatureData{
			SensorID:        sensorIDInt,
			Date:            ts,
			SoilTemperature: reading.Value,
			// ID might be set by DB or not needed if SDK handles it
		}

		// Construct the wrapper object for the SDK
		sdkPayload := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{tempDetail},
		}

		// Assuming SetSensorTemperatureData exists and takes int sensorID and the sdkPayload
		err = m.sdkClient.SetSensorTemperatureData(ctx, sensorIDInt, sdkPayload) // Pass the correctly constructed sdkPayload
		if err != nil {
			m.logger.Error().Err(err).Int("sensorID", sensorIDInt).Msg("Failed to set sensor temperature data via SDK")
			return
		}
		// Update the log to show the actual reading received
		m.logger.Info().Int("sensorID", sensorIDInt).Float64("Temperature", reading.Value).Int64("Timestamp", reading.Timestamp).Msg("Successfully processed and sent temperature data")

	case "4103": // Code for moisture
		var reading api.SensorReading
		err := json.Unmarshal(payload, &reading)

		if err != nil {

			m.logger.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal sensor reading")

			return

		}

		// Convert timestamp if needed (e.g., Unix ms to time.Time)

		ts := time.UnixMilli(reading.Timestamp) // Example conversion

		moisturePayloadForSDK := api.SensorMoistureData{

			SensorID: sensorIDInt,

			Date: ts,

			SoilMoisture: reading.Value,

			// ID might be set by DB or not needed here

		}

		// Adapt the SDK call based on what it expects.

		// Does it expect the single SensorMoistureData struct or the wrapper?

		// This is just a guess based on previous code:

		sdkResponseWrapper := api.SetSensorMoistureDataResponse{

			SensorMoistureData: []api.SensorMoistureData{moisturePayloadForSDK},
		}

		// The original SDK call used 'moistureData' which was SetSensorMoistureDataResponse

		// You need to adjust this call based on what m.sdkClient.SetSensorMoistureData actually accepts.

		// It might need the 'reading' directly, or the 'moisturePayloadForSDK', or the 'sdkResponseWrapper'.

		// Let's assume it needs the wrapper for now, like the original code did:

		err = m.sdkClient.SetSensorMoistureData(ctx, sensorIDInt, sdkResponseWrapper) // <-- Pass adapted data

		if err != nil {

			m.logger.Error().Err(err).Int("sensorID", sensorIDInt).Msg("Failed to set sensor moisture data via SDK")

			return

		}

		// Update the log to show the actual reading received

		m.logger.Info().Int("sensorID", sensorIDInt).Float64("Moisture", reading.Value).Int64("Timestamp", reading.Timestamp).Msg("Successfully processed and sent moisture data")

	default:

		m.logger.Warn().Str("topic", topic).Str("valueIdentifier", valueIdentifier).Msg("Received message with unhandled value identifier")

	}

}
