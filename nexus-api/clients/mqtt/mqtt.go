package clients

import (
	"context"
	"fmt"
	"time"

	"nexus-api/logging"

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
}

// MQTTClient wraps a connection to an MQTT broker
type MQTTClient struct {
	client mqtt.Client
	logger *logging.ServiceLogger
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
