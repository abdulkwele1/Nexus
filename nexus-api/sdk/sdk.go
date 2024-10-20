// package sdk implements clients for programmatically
// interacting with the REST APIs of the relay services
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"nexus-api/api"
	"nexus-api/logging"
)

const (
	HealthCheckPath    = "/healthcheck"
	LoginPath          = "/login"
	ChangePasswordPath = "/change-password"
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
	Config SDKConfig
	*logging.ServiceLogger
	Cookie *http.Cookie
}

// ReceiverServiceHealthCheck calls the service check endpoint
// for the Nexus API, returning error for any response with
// a non-200 status code
func (nc *NexusClient) HealthCheck(ctx context.Context) error {
	request, err := http.NewRequest("GET", nc.Config.NexusAPIEndpoint+HealthCheckPath, nil)

	if err != nil {
		return err
	}

	response, err := nc.http.Do(request)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}

// Login attempts to login to nexus app using username and password
// returning error (if any)
func (nc *NexusClient) Login(ctx context.Context, params api.LoginRequest) (api.LoginResponse, error) {
	body, err := json.Marshal(&params)

	if err != nil {
		return api.LoginResponse{}, err
	}

	request, err := http.NewRequest("POST", nc.Config.NexusAPIEndpoint+LoginPath, bytes.NewBuffer(body))

	if err != nil {
		return api.LoginResponse{}, err
	}

	nc.Trace().Msgf("sending request with params %+v\n headers %+v", params, request.Header)
	response, err := nc.http.Do(request)

	if err != nil {
		return api.LoginResponse{}, err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return api.LoginResponse{}, fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	var result api.LoginResponse
	err = json.NewDecoder(response.Body).Decode(&result)

	if err != nil {
		return api.LoginResponse{}, err
	}
	// save cookie for client
	nc.Cookie.Value = result.Cookie
	nc.Cookie.Name = "session_id"

	return result, nil
}

// ChangePassword attempts to change the users password
// returning error (if any)
func (nc *NexusClient) ChangePassword(ctx context.Context, params api.ChangePasswordRequest) error {
	body, err := json.Marshal(&params)

	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", nc.Config.NexusAPIEndpoint+ChangePasswordPath, bytes.NewBuffer(body))

	if err != nil {
		return err
	}

	nc.Trace()
	err = SetAuthHeaders(request, nc.Cookie)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("sending request with params %+v\n headers %+v", params, request.Header)
	response, err := nc.http.Do(request)

	if err != nil {
		return err
	}

	nc.Trace().Msgf("response %+v", response)

	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return fmt.Errorf("non 200 level status code %d", response.StatusCode)
	}

	return nil
}

// SetAuthHeaders sets the headers needed to authenticate requests
// to the Nexus API, returning error (if any)
func SetAuthHeaders(request *http.Request, cookie *http.Cookie) error {
	request.AddCookie(cookie)
	return nil
}

// NewClient creates a new client using the provided configuration
// returning the client and error (if any)
func NewClient(config SDKConfig) (*NexusClient, error) {
	client := NexusClient{
		http:          http.Client{},
		Config:        config,
		ServiceLogger: config.Logger,
		Cookie:        &http.Cookie{},
	}

	return &client, nil
}
