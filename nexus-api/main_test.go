package main

import (
	"context"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/logging"
	"nexus-api/password"
	"nexus-api/sdk"
	"os"
	"testing"

	"github.com/google/uuid"
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

	nexusClientGenerator = func() *sdk.NexusClient {
		client, err := sdk.NewClient(sdk.SDKConfig{
			NexusAPIEndpoint: os.Getenv("TEST_NEXUS_API_URL"),
			UserName:         os.Getenv("TEST_NEXUS_SDK_USER_NAME"),
			Password:         os.Getenv("TEST_NEXUS_SDK_PASSWORD"),
			Logger:           &testLogger,
		})

		if err != nil {
			panic(err)
		}

		return client
	}

	databaseClient = func() *database.PostgresClient {
		client, err := database.NewPostgresClient(database.PostgresDatabaseConfig{
			DatabaseName:          os.Getenv("TEST_DATABASE_NAME"),
			DatabaseEndpointURL:   os.Getenv("TEST_DATABASE_ENDPOINT_URL"),
			DatabaseUsername:      os.Getenv("TEST_DATABASE_USERNAME"),
			DatabasePassword:      os.Getenv("TEST_DATABASE_PASSWORD"),
			SSLEnabled:            false,
			QueryLoggingEnabled:   false,
			RunDatabaseMigrations: false,
			Logger:                &testLogger,
		})

		if err != nil {
			panic(err)
		}

		return &client
	}()
)

func TestE2ETestHealthCheckReturns200(t *testing.T) {
	// prepare test data
	testClient := nexusClientGenerator()
	err := testClient.HealthCheck(testCtx)

	assert.NoError(t, err)
}

func TestE2ETestLoginWithValidCredentialsReturnsCookie(t *testing.T) {
	// prepare test data
	testClient := nexusClientGenerator()
	// generate user login info
	testUserName := uuid.NewString()
	testPassword := uuid.NewString()

	testPasswordHash, err := password.HashPassword(testPassword)
	assert.NoError(t, err)
	// add user to database
	testLoginAuthentication := database.LoginAuthentication{
		UserName:     testUserName,
		PasswordHash: testPasswordHash,
	}

	err = testLoginAuthentication.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)
	// update test client to have credentials for test user
	testClient.Config.UserName = testUserName
	testClient.Config.Password = testPassword

	// execute test
	response, err := testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	// assert
	assert.NoError(t, err)
	// test the login info was correct
	assert.True(t, response.Match)
	// test a non-empty cookie was sent back
	assert.NotEqual(t, "", response.Cookie)
}
