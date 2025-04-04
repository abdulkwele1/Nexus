package clients

import (
	"context"
	"testing"
	"time"

	"nexus-api/logging"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
)

func TestNewMQTTClient(t *testing.T) {
	// Skip this test if no MQTT broker is available
	t.Skip("Skipping MQTT test as it requires a running MQTT broker")

	logger, err := logging.New("DEBUG")
	assert.NoError(t, err)

	config := MQTTConfig{
		BrokerURL:      "tcp://localhost:1883",
		ClientID:       "test-client-" + time.Now().Format("20060102150405"),
		CleanSession:   true,
		AutoReconnect:  true,
		ConnectTimeout: 5 * time.Second,
		Logger:         &logger,
	}

	client, err := NewMQTTClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.True(t, client.IsConnected())

	// Test subscription
	messageReceived := make(chan bool)
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		messageReceived <- true
	}

	err = client.Subscribe(context.Background(), "test/topic", 1, messageHandler)
	assert.NoError(t, err)

	// Test publishing
	err = client.Publish(context.Background(), "test/topic", 1, false, "test message")
	assert.NoError(t, err)

	// Wait for message to be received or timeout
	select {
	case <-messageReceived:
		// Message received successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Test unsubscription
	err = client.Unsubscribe(context.Background(), "test/topic")
	assert.NoError(t, err)

	// Test health check
	err = client.HealthCheck()
	assert.NoError(t, err)

	// Test disconnect
	client.Disconnect()
	assert.False(t, client.IsConnected())
}
