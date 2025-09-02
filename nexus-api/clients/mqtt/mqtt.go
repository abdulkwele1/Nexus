package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	client    mqtt.Client
	logger    *logging.ServiceLogger
	sdkClient *sdk.NexusClient

	// Track subscriptions for automatic resubscription on reconnect
	subscriptions     map[string]subscriptionInfo
	subscriptionMutex sync.RWMutex
}

// subscriptionInfo tracks subscription details for resubscription
type subscriptionInfo struct {
	topic    string
	qos      byte
	callback mqtt.MessageHandler
}

// NewMQTTClient returns a new connection to the specified MQTT broker and error (if any)

func NewMQTTClient(config MQTTConfig) (*MQTTClient, error) {

	// Set default values if not provided

	if config.ConnectTimeout == 0 {

		config.ConnectTimeout = 10 * time.Second

	}

	// Create MQTT client instance first to reference in handlers
	mqttClient := &MQTTClient{
		logger:        config.Logger,
		sdkClient:     config.SDKClient,
		subscriptions: make(map[string]subscriptionInfo),
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

			// Resubscribe to all tracked subscriptions
			mqttClient.resubscribeAll()

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

	// Set the client in the MQTT client instance
	mqttClient.client = client

	return mqttClient, nil

}

// resubscribeAll resubscribes to all tracked subscriptions
func (m *MQTTClient) resubscribeAll() {
	m.subscriptionMutex.RLock()
	defer m.subscriptionMutex.RUnlock()

	for topic, info := range m.subscriptions {
		token := m.client.Subscribe(topic, info.qos, info.callback)
		if token.Wait() && token.Error() != nil {
			m.logger.Error().Err(token.Error()).Msgf("Failed to resubscribe to topic: %s", topic)
		} else {
			m.logger.Info().Msgf("Resubscribed to topic: %s", topic)
		}
	}
}

// Subscribe subscribes to the specified topic with the given QoS level

// and message handler

func (m *MQTTClient) Subscribe(ctx context.Context, topic string, qos byte, callback mqtt.MessageHandler) error {

	// Track the subscription for automatic resubscription
	m.subscriptionMutex.Lock()
	m.subscriptions[topic] = subscriptionInfo{
		topic:    topic,
		qos:      qos,
		callback: callback,
	}
	m.subscriptionMutex.Unlock()

	token := m.client.Subscribe(topic, qos, callback)

	if token.Wait() && token.Error() != nil {

		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())

	}

	m.logger.Info().Msgf("Subscribed to topic: %s", topic)

	return nil

}

// Unsubscribe unsubscribes from the specified topic

func (m *MQTTClient) Unsubscribe(ctx context.Context, topic string) error {

	// Remove from tracked subscriptions
	m.subscriptionMutex.Lock()
	delete(m.subscriptions, topic)
	m.subscriptionMutex.Unlock()

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

// SensorReading represents a single sensor reading
type SensorReading struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

// HandleMessage implements mqtt.MessageHandler
func (m *MQTTClient) HandleMessage(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()
	ctx := context.Background()

	m.logger.Info().Str("topic", topic).Msg("Received message")
	m.logger.Trace().Str("topic", topic).Str("payload", string(payload)).Msg("Raw MQTT Payload")

	// Function to handle auth refresh and retry
	retryWithRefresh := func(operation func() error) error {
		err := operation()
		if err != nil && strings.Contains(err.Error(), "401") {
			m.logger.Info().Msg("Auth failed, attempting to re-login SDK client...")

			loginParams := api.LoginRequest{
				Username: m.sdkClient.Config.UserName,
				Password: m.sdkClient.Config.Password,
			}

			if _, err := m.sdkClient.Login(ctx, loginParams); err != nil {
				m.logger.Error().Err(err).Msg("Failed to re-login SDK client")
				return err
			}

			// Retry the operation with new cookie
			return operation()
		}
		return err
	}

	// Parse topic parts
	parts := strings.Split(topic, "/")
	if parts[0] != "" {
		m.logger.Warn().Str("topic", topic).Msg("Received message on unexpected topic format")
		return
	}

	// Extract common information
	deviceID := parts[2]
	sensorID := parts[3]

	// Handle different topic types
	var valueIdentifier string
	if parts[1] == "device_sensor_data" {
		// Handle sensor data topic
		if len(parts) < 7 {
			m.logger.Warn().Str("topic", topic).Msg("Invalid sensor data topic format")
			return
		}
		valueIdentifier = parts[6]
	} else if parts[1] == "device_status_event" {
		// Handle device status topic format: /device_status_event/<OrgID>/<DeviceEUI>/<Reserved>/<StatusID>
		if len(parts) < 6 {
			m.logger.Warn().Str("topic", topic).Msg("Invalid device status topic format")
			return
		}

		// Verify this is a battery status message (StatusID = 3000)
		statusID := parts[5]
		if statusID != "3000" {
			m.logger.Warn().Str("topic", topic).Str("statusID", statusID).Msg("Unexpected status ID for device status event")
			return
		}

		// Log raw payload for debugging
		m.logger.Debug().
			Str("topic", topic).
			Str("payload", string(payload)).
			Msg("Received battery status payload")

		// Try to parse payload as standard sensor reading first
		var reading SensorReading
		if err := json.Unmarshal(payload, &reading); err != nil {
			// If that fails, try the battery_level format
			var statusPayload struct {
				Value     string `json:"value"`
				Timestamp int64  `json:"timestamp"`
			}
			if err := json.Unmarshal(payload, &statusPayload); err != nil {
				m.logger.Error().Err(err).
					Str("payload", string(payload)).
					Msg("Failed to parse battery status payload in either format")
				return
			}
			// Convert string value to float64
			value, err := strconv.ParseFloat(statusPayload.Value, 64)
			if err != nil {
				m.logger.Error().Err(err).
					Str("value", statusPayload.Value).
					Msg("Failed to parse battery value as float")
				return
			}
			reading = SensorReading{
				Value:     value,
				Timestamp: statusPayload.Timestamp,
			}
		}

		// Create standardized sensor reading format
		var err error
		payload, err = json.Marshal(reading)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to marshal battery level reading")
			return
		}
		valueIdentifier = "3000" // Battery level identifier
	} else {
		m.logger.Warn().Str("topic", topic).Msg("Unknown topic type")
		return
	}

	// Log the received topic parts for debugging
	m.logger.Debug().
		Str("deviceID", deviceID).
		Str("sensorID", sensorID).
		Str("valueIdentifier", valueIdentifier).
		Msg("Processing MQTT message")

	// Parse the payload
	var reading SensorReading
	if err := json.Unmarshal(payload, &reading); err != nil {
		m.logger.Error().Err(err).Msg("Failed to parse sensor reading payload")
		return
	}

	// Create timestamp
	ts := time.UnixMilli(reading.Timestamp)

	// Log the sensor data with identification
	m.logger.Info().
		Str("deviceID", deviceID).
		Str("sensorID", sensorID).
		Str("type", valueIdentifier).
		Float64("value", reading.Value).
		Time("timestamp", ts).
		Msg("Received sensor reading")

	// Process based on sensor type
	switch valueIdentifier {
	case "3000": // Battery Level
		batteryDetail := api.BatteryLevelData{
			Date:         ts,
			BatteryLevel: reading.Value,
		}
		sdkPayload := api.SetBatteryLevelDataResponse{
			BatteryLevelData: []api.BatteryLevelData{batteryDetail},
		}
		err := retryWithRefresh(func() error {
			return m.sdkClient.SetSensorBatteryData(ctx, sensorID, sdkPayload)
		})
		if err != nil {
			m.logger.Error().Err(err).
				Str("deviceID", deviceID).
				Str("sensorID", sensorID).
				Msg("Failed to set battery data")
			return
		}
		m.logger.Info().
			Str("deviceID", deviceID).
			Str("sensorID", sensorID).
			Float64("batteryLevel", reading.Value).
			Msg("Successfully processed battery data")

	case "4102": // Temperature
		tempDetail := api.SensorTemperatureData{
			SensorID:        sensorID, // Use hex ID directly
			Date:            ts,
			SoilTemperature: reading.Value,
		}
		sdkPayload := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{tempDetail},
		}
		err := retryWithRefresh(func() error {
			return m.sdkClient.SetSensorTemperatureData(ctx, sensorID, sdkPayload)
		})
		if err != nil {
			m.logger.Error().Err(err).
				Str("deviceID", deviceID).
				Str("sensorID", sensorID).
				Msg("Failed to set temperature data")
			return
		}
		m.logger.Info().
			Str("deviceID", deviceID).
			Str("sensorID", sensorID).
			Float64("temperature", reading.Value).
			Msg("Successfully processed temperature data")

	case "4103": // Moisture
		moistureDetail := api.SensorMoistureData{
			SensorID:     sensorID, // Use hex ID directly
			Date:         ts,
			SoilMoisture: reading.Value,
		}
		sdkPayload := api.SetSensorMoistureDataResponse{
			SensorMoistureData: []api.SensorMoistureData{moistureDetail},
		}
		err := retryWithRefresh(func() error {
			return m.sdkClient.SetSensorMoistureData(ctx, sensorID, sdkPayload)
		})
		if err != nil {
			m.logger.Error().Err(err).
				Str("deviceID", deviceID).
				Str("sensorID", sensorID).
				Msg("Failed to set moisture data")
			return
		}
		m.logger.Info().
			Str("deviceID", deviceID).
			Str("sensorID", sensorID).
			Float64("moisture", reading.Value).
			Msg("Successfully processed moisture data")

	default:
		m.logger.Warn().
			Str("deviceID", deviceID).
			Str("sensorID", sensorID).
			Str("valueIdentifier", valueIdentifier).
			Msg("Received message with unhandled value identifier")
	}
}
