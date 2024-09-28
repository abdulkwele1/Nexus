// package sdk implements clients for programmatically
// interacting with the REST APIs of the relay services
package sdk

import (
	"context"
	"fmt"
	"net/http"

	"nexus-api/logging"
)

const (
	HealthCheckPath = "/healthcheck"
)

// SDKConfig wraps values for configuring
// a NexusDK client
type SDKConfig struct {
	NexusAPIEndpoint string
	UserName         string
	Password         string
	Logger           *logging.ServiceLogger
}

// NexusClient is a client for making requests to
// the Nexus API
type NexusClient struct {
	http   http.Client
	config SDKConfig
	*logging.ServiceLogger
	Cookie *http.Cookie
}

// ReceiverServiceHealthCheck calls the service check endpoint
// for the Nexus API, returning error for any response with
// a non-200 status code
func (c *NexusClient) HealthCheck(ctx context.Context) error {

	request, err := http.NewRequest("GET", c.config.NexusAPIEndpoint+HealthCheckPath, nil)

	if err != nil {
		return err
	}

	response, err := c.http.Do(request)

	if err != nil {
		return err
	}

	c.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}
// NewClient creates a new client using the provided configuration
// returning the client and error (if any)
func NewClient(config SDKConfig) (*NexusClient, error) {
	client := NexusClient{
		http:          http.Client{},
		config:        config,
		ServiceLogger: config.Logger,
	}

	return &client, nil
}
