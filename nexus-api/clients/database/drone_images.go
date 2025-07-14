package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var (
	ErrorNoDroneImage = errors.New("no drone image found")
)

type DroneImage struct {
	ID          uuid.UUID              `bun:"id,pk"`
	FileName    string                 `bun:"file_name"`
	FilePath    string                 `bun:"file_path"`
	UploadDate  time.Time              `bun:"upload_date"`
	FileSize    int64                  `bun:"file_size"`
	MimeType    string                 `bun:"mime_type"`
	Description string                 `bun:"description"`
	Metadata    map[string]interface{} `bun:"metadata,type:jsonb"`
}

func (di *DroneImage) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(di).Exec(ctx)
	return err
}

func GetDroneImageByID(ctx context.Context, db *bun.DB, id uuid.UUID) (*DroneImage, error) {
	var image DroneImage
	err := db.NewSelect().Model(&image).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNoDroneImage
		}
		return nil, err
	}
	return &image, nil
}

func GetDroneImagesByDateRange(ctx context.Context, db *bun.DB, startDate, endDate time.Time) ([]DroneImage, error) {
	var images []DroneImage
	err := db.NewSelect().
		Model(&images).
		Where("upload_date >= ? AND upload_date <= ?", startDate, endDate).
		Order("upload_date DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return images, nil
}

func DeleteDroneImage(ctx context.Context, db *bun.DB, id uuid.UUID) error {
	result, err := db.NewDelete().Model((*DroneImage)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrorNoDroneImage
	}

	return nil
}
