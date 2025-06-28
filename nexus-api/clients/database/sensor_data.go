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
)

type SensorMoistureData struct {
	ID           int       `bun:"id,pk,autoincrement"`
	SensorID     int       `bun:"sensor_id"`
	Date         time.Time `bun:"date,notnull"`
	SoilMoisture float64   `bun:"soil_moisture"`
}

type SensorTemperatureData struct {
	ID              int       `bun:"id,pk,autoincrement"`
	SensorID        int       `bun:"sensor_id"`
	Date            time.Time `bun:"date"`
	SoilTemperature float64   `bun:"soil_temperature"`
}

type SensorCoordinates struct {
	Latitude  float64 `bun:"latitude"`
	Longitude float64 `bun:"longitude"`
}

type Sensor struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Location          string    `json:"location"`
	InstallationDate  time.Time `json:"installation_date"`
	SensorCoordinates `json:"sensor_coordinates"`
}

func GetSensorMoistureDataForSensorID(ctx context.Context, db *bun.DB, sensorID int) ([]SensorMoistureData, error) {
	log.Info().Msgf("[sensor_data.go] Querying moisture data for sensor_id: %d", sensorID)
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

func GetSensorTemperatureDataForSensorID(ctx context.Context, db *bun.DB, sensorID int) ([]SensorTemperatureData, error) {
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
func EnsureSensorExists(ctx context.Context, db *bun.DB, sensorID int, deviceID string) error {
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
		Name:             fmt.Sprintf("Sensor %X (Auto-created)", sensorID),
		Location:         fmt.Sprintf("Device %s", deviceID),
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

func GetSensorTemperatureData(ctx context.Context, db *bun.DB, sensorID int) ([]SensorTemperatureData, error) {
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

func CreateSensor(ctx context.Context, db *bun.DB, name, location string) (Sensor, error) {
	sensor := Sensor{
		Name:             name,
		Location:         location,
		InstallationDate: time.Now(),
	}
	_, err := db.NewInsert().Model(&sensor).Exec(ctx)
	return sensor, err
}

func GetSensorByID(ctx context.Context, db *bun.DB, id int) (Sensor, error) {
	var sensor Sensor
	err := db.NewSelect().Model(&sensor).Where("id = ?", id).Scan(ctx)
	return sensor, err
}
