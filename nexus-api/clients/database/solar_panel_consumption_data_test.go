package database

import (
	"context"
	"time"

	"nexus-api/logging"
	"os"
	"testing"

	"math/rand"

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

	databaseClient *PostgresClient
)

func init() {
	// Set test environment variables for Docker environment
	os.Setenv("TEST_DATABASE_ENDPOINT_URL", "localhost:5432")
	os.Setenv("TEST_DATABASE_NAME", "postgres")
	os.Setenv("TEST_DATABASE_USERNAME", "postgres")
	os.Setenv("TEST_DATABASE_PASSWORD", "password")

	// Initialize database client after environment variables are set
	client, err := NewPostgresClient(PostgresDatabaseConfig{
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

	databaseClient = &client
}

func TestE2ESaveAndGetConsumptionDataForPanelID(t *testing.T) {
	// Step 1: Create test consumption data for a panel
	// Panel ID to test
	testPanelID := rand.Intn(10000000)

	testSolarPanelConsumptionData := []SolarPanelConsumptionData{
		{
			ConsumedKwh: rand.Float64(),
			CapacityKwh: rand.Float64(),
			PanelID:     testPanelID,
			Date:        time.Now().Add(1 * time.Second).UTC(),
		},
		{
			ConsumedKwh: rand.Float64(),
			CapacityKwh: rand.Float64(),
			PanelID:     testPanelID,
			Date:        time.Now().UTC(),
		},
	}

	// Step 2: Save panel test data to database
	for _, solarPanelConsumptionData := range testSolarPanelConsumptionData {
		err := solarPanelConsumptionData.Save(testCtx, databaseClient.DB)

		assert.NoError(t, err)
	}
	// Step 3: Retrieve consumption data for the same panel from database
	savedSolarPanelConsumptionData, err := GetConsumptionDataForPanelID(testCtx, databaseClient.DB, testPanelID)

	assert.NoError(t, err)

	// Step 4: test that the data returned from the database matches what we saved to the database
	assert.Equal(t, len(testSolarPanelConsumptionData), len(savedSolarPanelConsumptionData))

	var matchedOne bool
	var matchedTwo bool

	firstTestSolarPanelConsumptionData := testSolarPanelConsumptionData[0]
	secondTestSolarPanelConsumptionData := testSolarPanelConsumptionData[1]

	for _, savedSolarPanelConsumptionDataRow := range savedSolarPanelConsumptionData {
		// test if the first testSolarPanelConsumptionData is equal to this row
		if savedSolarPanelConsumptionDataRow.Date == firstTestSolarPanelConsumptionData.Date && savedSolarPanelConsumptionDataRow.ConsumedKwh == firstTestSolarPanelConsumptionData.ConsumedKwh && savedSolarPanelConsumptionDataRow.CapacityKwh == firstTestSolarPanelConsumptionData.CapacityKwh && savedSolarPanelConsumptionDataRow.PanelID == firstTestSolarPanelConsumptionData.PanelID {
			matchedOne = true
			continue
		}
		// test if the second testSolarPanelConsumptionData is equal to this row
		if savedSolarPanelConsumptionDataRow.Date == secondTestSolarPanelConsumptionData.Date && savedSolarPanelConsumptionDataRow.ConsumedKwh == secondTestSolarPanelConsumptionData.ConsumedKwh && savedSolarPanelConsumptionDataRow.CapacityKwh == secondTestSolarPanelConsumptionData.CapacityKwh && savedSolarPanelConsumptionDataRow.PanelID == secondTestSolarPanelConsumptionData.PanelID {
			matchedTwo = true
			continue
		}
	}

	assert.True(t, matchedOne)
	assert.True(t, matchedTwo)
}
