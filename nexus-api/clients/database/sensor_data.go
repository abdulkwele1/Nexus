package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

var (
	ErrorNoSensorMoistureData    = errors.New("no sensor moisture data found")
	ErrorNoSensorTemperatureData = errors.New("no sensor temperature data found")
	ErrorNoSensorBatteryData     = errors.New("no sensor battery data found")
)

type SensorMoistureData struct {
	ID           int       `bun:"id,pk,autoincrement"`
	SensorID     string    `bun:"sensor_id"`
	Date         time.Time `bun:"date,notnull"`
	SoilMoisture float64   `bun:"soil_moisture"`
}

type SensorTemperatureData struct {
	ID              int       `bun:"id,pk,autoincrement"`
	SensorID        string    `bun:"sensor_id"`
	Date            time.Time `bun:"date"`
	SoilTemperature float64   `bun:"soil_temperature"`
}

type SensorCoordinates struct {
	Latitude  float64 `bun:"latitude"`
	Longitude float64 `bun:"longitude"`
}

type Sensor struct {
	ID               string    `json:"id" bun:"id,pk"`
	Name             string    `json:"name" bun:"name"`
	Location         string    `json:"location" bun:"location"`
	InstallationDate time.Time `json:"installation_date" bun:"installation_date"`
	SensorCoordinates
}

func GetSensorMoistureDataForSensorID(ctx context.Context, db *bun.DB, sensorID string) ([]SensorMoistureData, error) {
	log.Info().Msgf("[sensor_data.go] Querying moisture data for sensor_id: %s", sensorID)
	var data []SensorMoistureData
	err := db.NewSelect().
		Model(&data).
		Where("sensor_id = ?", sensorID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrorNoSensorMoistureData
	}

	return data, nil
}

func GetSensorTemperatureDataForSensorID(ctx context.Context, db *bun.DB, sensorID string) ([]SensorTemperatureData, error) {
	var data []SensorTemperatureData
	err := db.NewSelect().
		Model(&data).
		Where("sensor_id = ?", sensorID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrorNoSensorTemperatureData
	}

	return data, nil
}

func (d *SensorMoistureData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().
		Model(d).
		Exec(ctx)
	return err
}

func (d *SensorTemperatureData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().
		Model(d).
		Exec(ctx)
	return err
}

// EnsureSensorExists ensures a sensor exists in the database, creating it if it doesn't exist
// This function will auto-create sensors without checking if they're online
func EnsureSensorExists(ctx context.Context, db *bun.DB, sensorID string, deviceID string) error {
	// Check if sensor already exists
	var existingSensor Sensor
	err := db.NewSelect().
		Model(&existingSensor).
		Where("id = ?", sensorID).
		Scan(ctx)

	// If sensor exists, we're done
	if err == nil {
		return nil
	}

	// If error is not "no rows found", return the error
	if err.Error() != "sql: no rows in result set" {
		return err
	}

	// Create new sensor entry
	newSensor := Sensor{
		ID:               sensorID,
		Name:             "Sensor " + sensorID + " (Auto-created)",
		Location:         "Device " + deviceID,
		InstallationDate: time.Now(),
	}

	_, err = db.NewInsert().
		Model(&newSensor).
		Exec(ctx)

	return err
}

// EnsureSensorExistsIfOnline ensures a sensor exists in the database, but only creates it
// if the provided data timestamp indicates the sensor is online (data is recent).
// dataTimestamp: The timestamp of the incoming sensor data
// onlineThresholdHours: How many hours back we consider data to be "recent" (default: 24 hours)
func EnsureSensorExistsIfOnline(ctx context.Context, db *bun.DB, sensorID string, deviceID string, dataTimestamp time.Time, onlineThresholdHours int) error {
	// Check if sensor already exists
	var existingSensor Sensor
	err := db.NewSelect().
		Model(&existingSensor).
		Where("id = ?", sensorID).
		Scan(ctx)

	// If sensor exists, we're done
	if err == nil {
		return nil
	}

	// If error is not "no rows found", return the error
	if err.Error() != "sql: no rows in result set" {
		return err
	}

	// Check if the data timestamp is recent (sensor is online)
	now := time.Now()
	threshold := time.Duration(onlineThresholdHours) * time.Hour
	timeDiff := now.Sub(dataTimestamp)

	// If data is too old, don't create the sensor (it's offline)
	if timeDiff > threshold {
		return fmt.Errorf("sensor %s appears offline: data timestamp (%v) is older than %d hours", sensorID, dataTimestamp, onlineThresholdHours)
	}

	// Data is recent, sensor is online - create new sensor entry
	newSensor := Sensor{
		ID:               sensorID,
		Name:             "Sensor " + sensorID + " (Auto-created)",
		Location:         "Device " + deviceID,
		InstallationDate: time.Now(),
	}

	_, err = db.NewInsert().
		Model(&newSensor).
		Exec(ctx)

	return err
}

func (spyd *Sensor) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(spyd).Exec(ctx)
	return err
}

func (c *PostgresClient) GetAllSensors(ctx context.Context, username string) ([]Sensor, error) {
	var sensors []Sensor
	// TODO: When users are associated with sensors, filter by username/userid
	err := c.DB.NewSelect().Model(&sensors).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return sensors, nil
}

func GetSensorTemperatureData(ctx context.Context, db *bun.DB, sensorID string) ([]SensorTemperatureData, error) {
	var data []SensorTemperatureData
	err := db.NewSelect().
		Model(&data).
		Where("sensor_id = ?", sensorID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrorNoSensorTemperatureData
	}

	return data, nil
}

func CreateSensor(ctx context.Context, db *bun.DB, eui, name, location string) (Sensor, error) {
	sensor := Sensor{
		ID:               eui,
		Name:             name,
		Location:         location,
		InstallationDate: time.Now(),
	}
	_, err := db.NewInsert().Model(&sensor).Exec(ctx)
	return sensor, err
}

func GetSensorByID(ctx context.Context, db *bun.DB, id string) (Sensor, error) {
	var sensor Sensor
	err := db.NewSelect().Model(&sensor).Where("id = ?", id).Scan(ctx)
	return sensor, err
}

func DeleteSensor(ctx context.Context, db *bun.DB, id string) error {
	// Start a transaction to ensure all deletions succeed or none do
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete all related sensor data first (due to foreign key constraints)
	// 1. Delete battery data
	_, err = tx.NewDelete().Model((*SensorBatteryData)(nil)).Where("sensor_id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	// 2. Delete moisture data
	_, err = tx.NewDelete().Model((*SensorMoistureData)(nil)).Where("sensor_id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	// 3. Delete temperature data
	_, err = tx.NewDelete().Model((*SensorTemperatureData)(nil)).Where("sensor_id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	// 4. Finally delete the sensor itself
	_, err = tx.NewDelete().Model((*Sensor)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

type SensorBatteryData struct {
	ID           int       `bun:"id,pk,autoincrement"`
	SensorID     string    `bun:"sensor_id"`
	Date         time.Time `bun:"date"`
	BatteryLevel float64   `bun:"battery_level"`
}

func GetSensorBatteryDataForSensorID(ctx context.Context, db *bun.DB, sensorID string, startDate, endDate time.Time) ([]SensorBatteryData, error) {
	var data []SensorBatteryData
	query := db.NewSelect().Model(&data).Where("sensor_id = ?", sensorID)

	if !startDate.IsZero() {
		query.Where("date >= ?", startDate)
	}
	if !endDate.IsZero() {
		query.Where("date <= ?", endDate)
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, ErrorNoSensorBatteryData
	}

	return data, nil
}

func (d *SensorBatteryData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(d).Exec(ctx)
	return err
}
