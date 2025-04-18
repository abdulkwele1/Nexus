package main

import (
	"context"
	"math/rand"
	"nexus-api/api"
	"nexus-api/clients/database"
	"nexus-api/logging"
	"nexus-api/password"
	"nexus-api/sdk"
	"os"
	"testing"
	"time"

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

	// test cookie is saved to database
	cookie, err := database.GetLoginCookie(testCtx, databaseClient.DB, response.Cookie)
	assert.NoError(t, err)
	assert.Equal(t, cookie.Cookie, response.Cookie)
}

func TestE2ETestChangePasswordAndLoginWithChangedPasswordSucceeds(t *testing.T) {
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

	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	newPassword := uuid.New().String()
	changePasswordParams := api.ChangePasswordRequest{
		CurrentPassword: testPassword,
		NewPassword:     newPassword,
	}

	// execute test
	// change password
	err = testClient.ChangePassword(testCtx, changePasswordParams)
	assert.NoError(t, err)

	// login with new password
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: newPassword,
	})

	// assert
	assert.NoError(t, err)
}

func TestE2ETestLogoutDeletesCookieFromDatabase(t *testing.T) {
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

	response, err := testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	cookie, err := database.GetLoginCookie(testCtx, databaseClient.DB, response.Cookie)
	assert.NoError(t, err)
	assert.Equal(t, cookie.Cookie, response.Cookie)

	// execute test
	err = testClient.Logout(testCtx)

	// assert
	assert.NoError(t, err)
	// test cookie is deleted from database
	_, err = database.GetLoginCookie(testCtx, databaseClient.DB, response.Cookie)
	assert.Error(t, err, "expected cookie to be deleted from database")
}

func TestE2ESetAndGetPanelYieldData(t *testing.T) {
	// Step: 0 prepare test data
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

	// login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	// Panel ID to test
	panelID := rand.Intn(10000000)

	// Test payload for setting yield data
	expectedYieldData := api.SetPanelYieldDataResponse{YieldData: []api.YieldData{
		{Date: time.Now().Add(1 * time.Second).UTC(), KwhYield: 100},
		{Date: time.Now().Add(1 * time.Second).UTC(), KwhYield: 150},
	}}

	// Step 1: POST (Set) yield data
	err = testClient.SetPanelYieldData(testCtx, panelID, expectedYieldData)
	assert.NoError(t, err, "Setting yield data should succeed")

	// Step 2: GET yield data
	gotYieldData, err := testClient.GetPanelYieldData(testCtx, panelID)
	assert.NoError(t, err, "Retrieving yield data should succeed")

	// Step 3: Compare
	assert.Equal(t, expectedYieldData.YieldData, gotYieldData.YieldData, "Yield data should match what was set")
}

func TestE2ESetAndGetPanelConsumptionData(t *testing.T) {
	// Step: 0 prepare test data
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

	// login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	// Panel ID to test
	panelID := rand.Intn(10000000)

	// Test payload for setting consumption data
	expectedConsumptionData := api.SetPanelConsumptionDataResponse{ConsumptionData: []api.ConsumptionData{
		{Date: time.Now().Add(1 * time.Second).UTC(), ConsumedKwh: 100},
		{Date: time.Now().Add(1 * time.Second).UTC(), CapacityKwh: 150},
	}}

	// Step 1: POST (Set) consumption data
	err = testClient.SetPanelConsumptionData(testCtx, panelID, expectedConsumptionData)
	assert.NoError(t, err, "Setting consumption data should succeed")

	// Step 2: GET consumption data
	gotConsumptionData, err := testClient.GetPanelConsumptionData(testCtx, panelID)
	assert.NoError(t, err, "Retrieving consumption data should succeed")

	// Step 3: Compare
	assert.Equal(t, expectedConsumptionData.ConsumptionData, gotConsumptionData.ConsumptionData, "Consumption data should match what was set")
}

func TestE2ESetAndGetSensorMoistureData(t *testing.T) {
	// Step: 0 prepare test data
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

	// login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	// sensor ID to test
	sensorID := rand.Intn(10000000)

	// add user to database
	testSensor := database.Sensor{
		ID: sensorID,
	}

	err = testSensor.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Test payload for setting Moisture data
	expectedMoistureData := api.SetSensorMoistureDataResponse{SensorMoistureData: []api.SensorMoistureData{
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilMoisture: 100, SensorID: sensorID},
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilMoisture: 150, SensorID: sensorID},
	}}

	// Step 1: POST (Set) moisture data
	err = testClient.SetSensorMoistureData(testCtx, sensorID, expectedMoistureData)
	assert.NoError(t, err, "Setting yield data should succeed")

	// change to get sensor moisture data
	gotMoistureData, err := testClient.GetSensorMoistureData(testCtx, sensorID)
	assert.NoError(t, err, "Retrieving yield data should succeed")

	// Step 3: compare to moisture data - ignoring ID field which is auto-generated
	assert.Equal(t, len(expectedMoistureData.SensorMoistureData), len(gotMoistureData.SensorMoistureData),
		"Number of moisture data entries should match")

	for i, expected := range expectedMoistureData.SensorMoistureData {
		actual := gotMoistureData.SensorMoistureData[i]
		assert.Equal(t, expected.SensorID, actual.SensorID, "SensorID should match")
		assert.Equal(t, expected.SoilMoisture, actual.SoilMoisture, "SoilMoisture should match")
		// Compare dates with a small tolerance to account for potential time differences
		assert.True(t, expected.Date.Sub(actual.Date) < time.Second,
			"Date should be within 1 second of expected")
	}
}

func TestE2ESetAndGetSensorTemperatureData(t *testing.T) {
	// Step: 0 prepare test data
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

	// login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})

	assert.NoError(t, err)

	// sensor ID to test
	sensorID := rand.Intn(10000000)

	// add user to database
	testSensor := database.Sensor{
		ID: sensorID,
	}

	err = testSensor.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Test payload for setting Moisture data
	expectedTemperatureData := api.SetSensorTemperatureDataResponse{SensorTemperatureData: []api.SensorTemperatureData{
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilTemperature: 100, SensorID: sensorID},
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilTemperature: 150, SensorID: sensorID},
	}}

	// Step 1: POST (Set) moisture data
	err = testClient.SetSensorTemperatureData(testCtx, sensorID, expectedTemperatureData)
	assert.NoError(t, err, "Setting Temperature data should succeed")

	// change to get sensor moisture data
	gotTemperatureData, err := testClient.GetSensorTemperatureData(testCtx, sensorID)
	assert.NoError(t, err, "Retrieving yield data should succeed")

	// Step 3: compare to moisture data - ignoring ID field which is auto-generated
	assert.Equal(t, len(expectedTemperatureData.SensorTemperatureData), len(gotTemperatureData.SensorTemperatureData),
		"Number of moisture data entries should match")

	for i, expected := range expectedTemperatureData.SensorTemperatureData {
		actual := gotTemperatureData.SensorTemperatureData[i]
		assert.Equal(t, expected.SensorID, actual.SensorID, "SensorID should match")
		assert.Equal(t, expected.SoilTemperature, actual.SoilTemperature, "SoilTemperature should match")
		// Compare dates with a small tolerance to account for potential time differences
		assert.True(t, expected.Date.Sub(actual.Date) < time.Second,
			"Date should be within 1 second of expected")
	}
}
