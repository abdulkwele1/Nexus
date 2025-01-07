package database

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrorNoSolarPanelYieldData = errors.New("no energy production data found for panel")
)

type SolarPanelYieldData struct {
	bun.BaseModel `bun:"table:solar_panel_yield_data"`
	ID            int       `bun:"id,pk,autoincrement"`
	Date          time.Time `bun:"date,notnull"`
	KwHYield      float64   `bun:"kwh_yield,notnull"`
	PanelID       int       `bun:"panel_id,notnull"`
}

// Save saves the current cookie to
// the database, returning error (if any)
func (spyd *SolarPanelYieldData) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(spyd).Exec(ctx)

	return err
}

// Upsert inserts or updates the cookie for the user
// in the database, returning error (if any)
func (spyd *SolarPanelYieldData) Upsert(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().On("CONFLICT (date) DO UPDATE").Model(spyd).Exec(ctx)

	return err
}

// Load returns the row in the login_cookies table
// returning error (if any)
func (spyd *SolarPanelYieldData) Load(ctx context.Context, db *bun.DB) error {
	return db.NewSelect().Model(spyd).WherePK().Scan(ctx)
}

// Update updates a row in the login_cookies table
// returning error (if any)
func (spyd *SolarPanelYieldData) Update(ctx context.Context, db *bun.DB) error {
	_, err := db.NewUpdate().Model(spyd).WherePK().Exec(ctx)

	return err
}

// Delete deletes the cookie for the username
// returning error (if any)
func (spyd *SolarPanelYieldData) Delete(ctx context.Context, db *bun.DB) error {
	_, err := db.NewDelete().Model(spyd).WherePK().Exec(ctx)

	return err
}

func GetDataForPanelID(ctx context.Context, db *bun.DB, panelID int) ([]SolarPanelYieldData, error) {
	var solarPanelData []SolarPanelYieldData
	err := db.NewSelect().Model(&solarPanelData).Where("panel_id = ?", panelID).Scan(ctx)

	if err != nil {
		return solarPanelData, err
	}

	if len(solarPanelData) == 0 {
		return solarPanelData, ErrorNoSolarPanelYieldData
	}

	return solarPanelData, nil
}

func DeleteDataFoPanelID(ctx context.Context, panelID string, db *bun.DB) error {
	_, err := db.NewDelete().Model(&SolarPanelYieldData{}).Where("panel_id = ?", panelID).Exec(ctx)

	return err
}
