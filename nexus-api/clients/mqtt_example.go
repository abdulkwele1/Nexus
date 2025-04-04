package clients

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nexus-api/logging"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ExampleMQTTClient demonstrates how to use the MQTT client
func ExampleMQTTClient() {
	// Create a logger
	logger, err := logging.New("DEBUG")
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		return
	}

	// Create MQTT client configuration
	config := MQTTConfig{
		BrokerURL:      "tcp://localhost:1883", // Replace with your MQTT broker URL
		ClientID:       "example-client-" + time.Now().Format("20060102150405"),
		Username:       "", // Replace with your username if required
		Password:       "", // Replace with your password if required
		CleanSession:   true,
		AutoReconnect:  true,
		ConnectTimeout: 10 * time.Second,
		Logger:         &logger,
	}

	// Create the MQTT client
	client, err := NewMQTTClient(config)
	if err != nil {
		fmt.Printf("Failed to create MQTT client: %v\n", err)
		return
	}
	defer client.Disconnect()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a signal handler to gracefully shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Define a message handler
	messageHandler := func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message on topic %s: %s\n", msg.Topic(), string(msg.Payload()))
	}

	// Subscribe to a topic
	topic := "example/topic"
	err = client.Subscribe(ctx, topic, 1, messageHandler)
	if err != nil {
		fmt.Printf("Failed to subscribe to topic %s: %v\n", topic, err)
		return
	}

	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Publish a message to the topic
	message := "Hello, MQTT!"
	err = client.Publish(ctx, topic, 1, false, message)
	if err != nil {
		fmt.Printf("Failed to publish message: %v\n", err)
		return
	}

	fmt.Printf("Published message: %s\n", message)

	// Wait for a signal to shut down
	<-sigChan
	fmt.Println("Shutting down...")
}
