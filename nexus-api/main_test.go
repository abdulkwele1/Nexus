package main

import (
	"context"
	"fmt"
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

	databaseClient *database.PostgresClient
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
	// Create a test sensor first
	sensor := &database.Sensor{
		Name:             fmt.Sprintf("Test Sensor %d", time.Now().Unix()),
		Location:         "Test Location",
		InstallationDate: time.Now(),
		SensorCoordinates: database.SensorCoordinates{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	// Save the sensor to the database
	err := sensor.Save(testCtx, databaseClient.DB)
	if err != nil {
		t.Fatalf("Failed to save sensor: %v", err)
	}

	if sensor.ID == 0 {
		t.Fatal("Sensor ID was not set after save")
	}

	// Use a unique timestamp for this test run
	testTime := time.Now().Add(time.Hour * 24) // Use tomorrow to avoid conflicts

	// Prepare test data
	testData := api.SetSensorMoistureDataResponse{
		SensorMoistureData: []api.SensorMoistureData{
			{
				SensorID:     sensor.ID,
				Date:         testTime.Format(time.RFC3339),
				SoilMoisture: 42.5,
			},
		},
	}

	// Generate user login info
	testUserName := uuid.NewString()
	testPassword := uuid.NewString()

	testPasswordHash, err := password.HashPassword(testPassword)
	assert.NoError(t, err)

	// Add user to database
	testLoginAuthentication := database.LoginAuthentication{
		UserName:     testUserName,
		PasswordHash: testPasswordHash,
	}

	err = testLoginAuthentication.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Create test client with credentials
	testClient := nexusClientGenerator()
	testClient.Config.UserName = testUserName
	testClient.Config.Password = testPassword

	// Login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Set moisture data
	err = testClient.SetSensorMoistureData(testCtx, sensor.ID, testData)
	if err != nil {
		t.Fatalf("Failed to set moisture data: %v", err)
	}

	// Get moisture data
	retrievedData, err := testClient.GetSensorMoistureData(testCtx, sensor.ID)
	if err != nil {
		t.Fatalf("Failed to get moisture data: %v", err)
	}

	// Verify the data
	if len(retrievedData.SensorMoistureData) != len(testData.SensorMoistureData) {
		t.Errorf("Expected %d records, got %d", len(testData.SensorMoistureData), len(retrievedData.SensorMoistureData))
	}

	for i, data := range retrievedData.SensorMoistureData {
		if data.SensorID != testData.SensorMoistureData[i].SensorID {
			t.Errorf("Expected SensorID %d, got %d", testData.SensorMoistureData[i].SensorID, data.SensorID)
		}
		if data.SoilMoisture != testData.SensorMoistureData[i].SoilMoisture {
			t.Errorf("Expected SoilMoisture %f, got %f", testData.SensorMoistureData[i].SoilMoisture, data.SoilMoisture)
		}
	}
}

func TestE2ESetAndGetSensorTemperatureData(t *testing.T) {
	// Create a test sensor first
	sensor := &database.Sensor{
		Name:             fmt.Sprintf("Test Sensor %d", time.Now().Unix()),
		Location:         "Test Location",
		InstallationDate: time.Now(),
		SensorCoordinates: database.SensorCoordinates{
			Latitude:  37.7749,
			Longitude: -122.4194,
		},
	}

	// Save the sensor to the database
	err := sensor.Save(testCtx, databaseClient.DB)
	if err != nil {
		t.Fatalf("Failed to save sensor: %v", err)
	}

	if sensor.ID == 0 {
		t.Fatal("Sensor ID was not set after save")
	}

	// Use a unique timestamp for this test run
	testTime := time.Now().Add(time.Hour * 48) // Use day after tomorrow to avoid conflicts

	// Prepare test data
	testData := api.SetSensorTemperatureDataResponse{
		SensorTemperatureData: []api.SensorTemperatureData{
			{
				SensorID:        sensor.ID,
				Date:            testTime.Format(time.RFC3339),
				SoilTemperature: 25.5,
			},
		},
	}

	// Generate user login info
	testUserName := uuid.NewString()
	testPassword := uuid.NewString()

	testPasswordHash, err := password.HashPassword(testPassword)
	assert.NoError(t, err)

	// Add user to database
	testLoginAuthentication := database.LoginAuthentication{
		UserName:     testUserName,
		PasswordHash: testPasswordHash,
	}

	err = testLoginAuthentication.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Create test client with credentials
	testClient := nexusClientGenerator()
	testClient.Config.UserName = testUserName
	testClient.Config.Password = testPassword

	// Login user
	_, err = testClient.Login(testCtx, api.LoginRequest{
		Username: testUserName,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	// Set temperature data
	err = testClient.SetSensorTemperatureData(testCtx, sensor.ID, testData)
	if err != nil {
		t.Fatalf("Failed to set temperature data: %v", err)
	}

	// Get temperature data
	retrievedData, err := testClient.GetSensorTemperatureData(testCtx, sensor.ID)
	if err != nil {
		t.Fatalf("Failed to get temperature data: %v", err)
	}

	// Verify the data
	if len(retrievedData.SensorTemperatureData) != len(testData.SensorTemperatureData) {
		t.Errorf("Expected %d records, got %d", len(testData.SensorTemperatureData), len(retrievedData.SensorTemperatureData))
	}

	for i, data := range retrievedData.SensorTemperatureData {
		if data.SensorID != testData.SensorTemperatureData[i].SensorID {
			t.Errorf("Expected SensorID %d, got %d", testData.SensorTemperatureData[i].SensorID, data.SensorID)
		}
		if data.SoilTemperature != testData.SensorTemperatureData[i].SoilTemperature {
			t.Errorf("Expected SoilTemperature %f, got %f", testData.SensorTemperatureData[i].SoilTemperature, data.SoilTemperature)
		}
	}
}
