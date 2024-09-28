package main

import (
	"context"
	"nexus-api/logging"
	"nexus-api/sdk"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testCtx    = context.TODO()
	testLogger = func() logging.ServiceLogger {
		logger, err := logging.New("INFO")

		if err != nil {
			panic(err)
		}

		return logger
	}()

	nexusClient = func() *sdk.NexusClient {
		client, err := sdk.NewClient(sdk.SDKConfig{
			NexusAPIEndpoint: os.Getenv("TEST_NEXUS_API_URL"),
			Logger:           &testLogger,
		})

		if err != nil {
			panic(err)
		}

		return client
	}()
)

func TestE2ETestHealthCheckReturns200(t *testing.T) {
	err := nexusClient.HealthCheck(testCtx)

	assert.NoError(t, err)
}
