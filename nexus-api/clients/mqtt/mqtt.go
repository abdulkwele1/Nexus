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

// SensorMessage is the message structure for sensor data
type SensorMessage struct {
	SensorID  string  `json:"sensor_id"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

// SDKMQTTHandler is the handler for SDK MQTT messages
type SDKMQTTHandler struct {
	sdkClient *sdk.NexusClient
	logger    *logging.ServiceLogger
}

// NewSDKMQTTHandler creates a new SDKMQTTHandler
func NewSDKMQTTHandler(sdkClient *sdk.NexusClient, logger *logging.ServiceLogger) *SDKMQTTHandler {
	return &SDKMQTTHandler{sdkClient: sdkClient, logger: logger}
}

func (h *SDKMQTTHandler) HandleSensorData(ctx context.Context, topic string, payload []byte) error {
	var msg SensorMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("parse sensor message: %w", err)
	}

	parts := strings.Split(topic, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid topic format: %s", topic)
	}

	sensorID := parts[2]
	sensorType := parts[len(parts)-1]
	h.logger.Info().Msgf("Sensor ID=%s Type=%s", sensorID, sensorType)

	idInt, err := strconv.Atoi(sensorID)
	if err != nil {
		h.logger.Warn().Err(err).
			Msgf("non‐numeric sensor ID %s, defaulting to 0", sensorID)
		idInt = 0
	}

	ts := time.UnixMilli(msg.Timestamp)

	switch sensorType {
	case "4102": // moisture
		payload := api.SetSensorMoistureDataResponse{
			SensorMoistureData: []api.SensorMoistureData{{
				Date:         ts,
				SoilMoisture: msg.Value,
			}},
		}
		if err := h.sdkClient.SetSensorMoistureData(ctx, idInt, payload); err != nil {
			return fmt.Errorf("save moisture: %w", err)
		}
		h.logger.Info().Msgf("Saved moisture for %s", sensorID)

	case "4103": // temperature
		payload := api.SetSensorTemperatureDataResponse{
			SensorTemperatureData: []api.SensorTemperatureData{{
				Date:            ts,
				SoilTemperature: msg.Value,
			}},
		}
		if err := h.sdkClient.SetSensorTemperatureData(ctx, idInt, payload); err != nil {
			return fmt.Errorf("save temperature: %w", err)
		}
		h.logger.Info().Msgf("Saved temperature for %s", sensorID)

	default:
		return fmt.Errorf("unknown sensor type %s", sensorType)
	}

	return nil
}

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
	client  mqtt.Client
	logger  *logging.ServiceLogger
	handler *SDKMQTTHandler
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

	handler := NewSDKMQTTHandler(config.SDKClient, config.Logger)

	return &MQTTClient{
		client:  client,
		logger:  config.Logger,
		handler: handler,
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

// HandleMessage implements mqtt.MessageHandler and has access to sdkClient and logger
func (m *MQTTClient) HandleMessage(_ mqtt.Client, msg mqtt.Message) {
	topic, payload := msg.Topic(), msg.Payload()
	m.logger.Info().Msgf("Received message on topic %s: %s", topic, string(payload))

	// 4️⃣ hand off to your parsing & SDK logic
	if err := m.handler.HandleSensorData(context.Background(), topic, payload); err != nil {
		m.logger.Error().
			Err(err).
			Msgf("failed to process sensor data on %s", topic)
	}
}
