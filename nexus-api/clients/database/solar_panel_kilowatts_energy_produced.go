package database

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrorNoSolarPanelKilowattsEnergyProducedData = errors.New("no energy production data found for panel")
)

type SolarPanelKilowattsEnergyProducedData struct {
	bun.BaseModel `bun:"table:SolarPanelKilowattsEnergyProducedData"`
	ID            int       `bun:"id,pk,autoincrement"`
	Date          time.Time `bun:"date,notnull"`
	Production    int       `bun:"production,notnull"`
	PanelID       string    `bun:"panel_id,notnull"`
}

// Save saves the current cookie to
// the database, returning error (if any)
func (spkepd *SolarPanelKilowattsEnergyProducedData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(spkepd).Exec(ctx)

	return err
}

// Upsert inserts or updates the cookie for the user
// in the database, returning error (if any)
func (spkepd *SolarPanelKilowattsEnergyProducedData) Upsert(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().On("CONFLICT (date) DO UPDATE").Model(spkepd).Exec(ctx)

	return err
}

// Load returns the row in the login_cookies table
// returning error (if any)
func (spkepd *SolarPanelKilowattsEnergyProducedData) Load(ctx context.Context, db *bun.DB) error {
	return db.NewSelect().Model(spkepd).WherePK().Scan(ctx)
}

// Update updates a row in the login_cookies table
// returning error (if any)
func (spkepd *SolarPanelKilowattsEnergyProducedData) Update(ctx context.Context, db *bun.DB) error {
	_, err := db.NewUpdate().Model(spkepd).WherePK().Exec(ctx)

	return err
}

// Delete deletes the cookie for the username
// returning error (if any)
func (spkepd *SolarPanelKilowattsEnergyProducedData) Delete(ctx context.Context, db *bun.DB) error {
	_, err := db.NewDelete().Model(spkepd).WherePK().Exec(ctx)

	return err
}


func GetDataForPanelID(ctx context.Context, db *bun.DB, panelID string) ([]SolarPanelKilowattsEnergyProducedData, error) {
	var solarPanelData []SolarPanelKilowattsEnergyProducedData
	err := db.NewSelect().Model(&solarPanelData).Where("panel_id = ?", panelID).Scan(ctx)

	if err != nil {
		return solarPanelData, err
	}

	if len(solarPanelData) > 0 {
		return solarPanelData, ErrorNoSolarPanelKilowattsEnergyProducedData
	}

	return solarPanelData, nil
}

func DeleteDataFoPanelID(ctx context.Context, panelID string, db *bun.DB) error {
	_, err := db.NewDelete().Model(&SolarPanelKilowattsEnergyProducedData{}).Where("panel_id = ?", panelID).Exec(ctx)

	return err
}
