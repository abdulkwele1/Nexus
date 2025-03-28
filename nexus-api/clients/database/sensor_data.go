package database

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrorNoSensorMoistureData    = errors.New("no sensor moisture data found")
	ErrorNoSensorTemperatureData = errors.New("no sensor temperature data found")
)

type SensorMoistureData struct {
	ID           int       `bun:"id,pk,autoincrement"`
	SensorID     int       `bun:"sensor_id"`
	Date         time.Time `bun:"date"`
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
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	InstallationDate  string `json:"installation_date"`
	SensorCoordinates `json:"sensor_coordinates"`
}

func GetSensorMoistureDataForSensorID(ctx context.Context, db *bun.DB, sensorID int) ([]SensorMoistureData, error) {
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
		On("CONFLICT (sensor_id, date) DO UPDATE").
		Set("soil_moisture = EXCLUDED.soil_moisture").
		Exec(ctx)
	return err
}

func (d *SensorTemperatureData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().
		Model(d).
		On("CONFLICT (sensor_id, date) DO UPDATE").
		Set("soil_temperature = EXCLUDED.soil_temperature").
		Exec(ctx)
	return err
}
