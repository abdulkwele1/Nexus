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
	sensorID := "2CF7F1C06270008D" // Using a hex string ID

	// add sensor to database
	testSensor := database.Sensor{
		ID:       sensorID,
		Name:     "Test Sensor",
		Location: "Test Location",
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
	sensorID := "2CF7F1C0627000BC" // Using a hex string ID

	// add sensor to database
	testSensor := database.Sensor{
		ID:       sensorID,
		Name:     "Test Sensor",
		Location: "Test Location",
	}

	err = testSensor.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Test payload for setting Temperature data
	expectedTemperatureData := api.SetSensorTemperatureDataResponse{SensorTemperatureData: []api.SensorTemperatureData{
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilTemperature: 100, SensorID: sensorID},
		{Date: time.Now().Add(1 * time.Second).UTC(), SoilTemperature: 150, SensorID: sensorID},
	}}

	// Step 1: POST (Set) temperature data
	err = testClient.SetSensorTemperatureData(testCtx, sensorID, expectedTemperatureData)
	assert.NoError(t, err, "Setting Temperature data should succeed")

	// change to get sensor temperature data
	gotTemperatureData, err := testClient.GetSensorTemperatureData(testCtx, sensorID)
	assert.NoError(t, err, "Retrieving temperature data should succeed")

	// Step 3: compare to temperature data - ignoring ID field which is auto-generated
	assert.Equal(t, len(expectedTemperatureData.SensorTemperatureData), len(gotTemperatureData.SensorTemperatureData),
		"Number of temperature data entries should match")

	for i, expected := range expectedTemperatureData.SensorTemperatureData {
		actual := gotTemperatureData.SensorTemperatureData[i]
		assert.Equal(t, expected.SensorID, actual.SensorID, "SensorID should match")
		assert.Equal(t, expected.SoilTemperature, actual.SoilTemperature, "SoilTemperature should match")
		// Compare dates with a small tolerance to account for potential time differences
		assert.True(t, expected.Date.Sub(actual.Date) < time.Second,
			"Date should be within 1 second of expected")
	}
}

func TestE2EGetAllSensors(t *testing.T) {
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

	// Create some test sensors
	testSensors := []database.Sensor{
		{
			ID:       "2CF7F1C06270008D",
			Name:     "Test Sensor 1",
			Location: "Test Location 1",
		},
		{
			ID:       "2CF7F1C0627000BC",
			Name:     "Test Sensor 2",
			Location: "Test Location 2",
		},
		{
			ID:       "2CF7F1C0627000C4",
			Name:     "Test Sensor 3",
			Location: "Test Location 3",
		},
	}

	// Save test sensors to database
	for _, sensor := range testSensors {
		err = sensor.Save(testCtx, databaseClient.DB)
		assert.NoError(t, err, "Saving test sensor should succeed")
	}

	// Step 1: GET all sensors
	gotSensors, err := testClient.GetAllSensors(testCtx)
	assert.NoError(t, err, "Retrieving all sensors should succeed")

	// Step 2: Verify that our test sensors are in the response
	assert.GreaterOrEqual(t, len(gotSensors), len(testSensors), "Should have at least as many sensors as we created")

	// Check that each of our test sensors is present in the response
	foundSensors := make(map[string]bool)
	for _, sensor := range gotSensors {
		foundSensors[sensor.ID] = true
	}

	for _, expectedSensor := range testSensors {
		assert.True(t, foundSensors[expectedSensor.ID], "Test sensor with ID %s should be found in response", expectedSensor.ID)
	}

	// Step 3: Verify that at least one sensor has the expected structure
	if len(gotSensors) > 0 {
		firstSensor := gotSensors[0]
		assert.NotEmpty(t, firstSensor.Name, "Sensor should have a name")
		assert.NotEmpty(t, firstSensor.Location, "Sensor should have a location")
		assert.NotZero(t, firstSensor.ID, "Sensor should have a non-zero ID")
	}
}

func TestE2ESetAndGetSensorBatteryData(t *testing.T) {
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

	// Generate unique sensor ID for test (using shorter format)
	sensorID := fmt.Sprintf("TS%s", uuid.NewString()[:8])

	// add sensor to database
	testSensor := database.Sensor{
		ID:               sensorID,
		Name:             "Test Battery Sensor",
		Location:         "Test Location",
		InstallationDate: time.Now(),
	}

	err = testSensor.Save(testCtx, databaseClient.DB)
	assert.NoError(t, err)

	// Test payload for setting battery data
	expectedBatteryData := api.SetBatteryLevelDataResponse{
		BatteryLevelData: []api.BatteryLevelData{
			{
				Date:         time.Now().Add(1 * time.Second).UTC(),
				BatteryLevel: 85.5,
			},
			{
				Date:         time.Now().Add(2 * time.Second).UTC(),
				BatteryLevel: 90.0,
			},
		},
	}

	// Step 1: POST (Set) battery data
	err = testClient.SetSensorBatteryData(testCtx, sensorID, expectedBatteryData)
	assert.NoError(t, err, "Setting battery data should succeed")

	// Step 2: GET battery data
	startDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02") // Yesterday
	endDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")    // Tomorrow
	gotBatteryData, err := testClient.GetSensorBatteryData(testCtx, sensorID, startDate, endDate)
	assert.NoError(t, err, "Retrieving battery data should succeed")

	// Step 3: Compare battery data
	assert.NotNil(t, gotBatteryData, "Response should not be nil")
	assert.NotNil(t, gotBatteryData.BatteryLevelData, "Battery data array should not be nil")

	assert.Equal(t, len(expectedBatteryData.BatteryLevelData), len(gotBatteryData.BatteryLevelData),
		"Number of battery data entries should match")

	if assert.GreaterOrEqual(t, len(gotBatteryData.BatteryLevelData), len(expectedBatteryData.BatteryLevelData),
		"Should have at least as many entries as we sent") {

		for i, expected := range expectedBatteryData.BatteryLevelData {
			actual := gotBatteryData.BatteryLevelData[i]
			assert.Equal(t, expected.BatteryLevel, actual.BatteryLevel, "Battery level should match") // Compare dates with a small tolerance to account for potential time differences
			assert.True(t, expected.Date.Sub(actual.Date) < time.Second,
				"Date should be within 1 second of expected")
		}
	}
}

func TestE2ESetAndGetDroneImages(t *testing.T) {
	// Step 0: prepare test data
	testClient := nexusClientGenerator()
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

	// Test data
	testDescription := "Test drone image"
	testFileName := "test_image.jpg"
	testFileContent := []byte("test image content")
	testMetadata := map[string]interface{}{
		"test_key":  "test_value",
		"mime_type": "image/jpeg", // Add explicit mime type
	}

	// Step 1: Test Upload
	uploadResponse, err := testClient.UploadDroneImages(
		testCtx,
		[][]byte{testFileContent},
		[]string{testFileName},
		testDescription,
		testMetadata,
	)
	assert.NoError(t, err, "Uploading drone image should succeed")
	assert.Equal(t, 1, len(uploadResponse.UploadedImages), "Should have uploaded one image")
	uploadedImage := uploadResponse.UploadedImages[0]

	// Step 2: Test Get Single Image
	retrievedImage, err := testClient.GetDroneImage(testCtx, uploadedImage.ID)
	assert.NoError(t, err, "Retrieving single drone image should succeed")
	assert.Equal(t, uploadedImage.ID, retrievedImage.ID, "Retrieved image ID should match uploaded image")
	assert.Equal(t, testFileName, retrievedImage.FileName, "Retrieved filename should match uploaded filename")
	assert.Equal(t, testDescription, retrievedImage.Description, "Retrieved description should match uploaded description")

	// Step 3: Test Get All Images
	endDate := time.Now().Add(time.Hour)   // Add buffer to ensure we're after upload time
	startDate := endDate.AddDate(0, -1, 0) // Last 30 days
	allImages, err := testClient.GetDroneImages(testCtx, startDate, endDate)
	assert.NoError(t, err, "Retrieving all drone images should succeed")
	assert.GreaterOrEqual(t, len(allImages.Images), 1, "Should have at least one image")

	// Find our uploaded image in the list
	found := false
	for _, img := range allImages.Images {
		if img.ID == uploadedImage.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Uploaded image should be found in list of all images")

	// Step 4: Test Delete Image
	err = testClient.DeleteDroneImage(testCtx, uploadedImage.ID)
	assert.NoError(t, err, "Deleting drone image should succeed")

	// Verify deletion
	_, err = testClient.GetDroneImage(testCtx, uploadedImage.ID)
	assert.Error(t, err, "Getting deleted image should return error")
}
