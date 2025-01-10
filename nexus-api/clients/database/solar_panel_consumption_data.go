package database

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrorNoSolarPanelConsumptionData = errors.New("no energy production data found for panel")
)

type SolarPanelConsumptionData struct {
	bun.BaseModel `bun:"table:solar_panel_consumption_data"`
	ID            int       `bun:"id,pk,autoincrement"`
	Date          time.Time `bun:"date,notnull"`
	ConsumedKwh   float64   `bun:"consumed_kwh,notnull"`
	CapacityKwh   float64   `bun:"capacity_kwh,notnull"`
	PanelID       int       `bun:"panel_id,notnull"`
}

// Save saves the current consumption data to
// the database, returning error (if any)
func (spcd *SolarPanelConsumptionData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(spcd).Exec(ctx)

	return err
}

func GetConsumptionDataForPanelID(ctx context.Context, db *bun.DB, panelID int) ([]SolarPanelConsumptionData, error) {
	var solarPanelData []SolarPanelConsumptionData
	err := db.NewSelect().Model(&solarPanelData).Where("panel_id = ?", panelID).Scan(ctx)

	if err != nil {
		return solarPanelData, err
	}

	if len(solarPanelData) == 0 {
		return solarPanelData, ErrorNoSolarPanelConsumptionData
	}

	return solarPanelData, nil
}
