package database

import (
	"context"
	"fmt"
	"time"

	"nexus-api/logging"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// Migrate sets up and runs all migrations in the migrations model
// that haven't been run on the database being used by the proxy service
// returning error (if any) and a list of migrations that have been
// run and any that were not
func Migrate(ctx context.Context, db *bun.DB, migrations migrate.Migrations, logger *logging.ServiceLogger) (*migrate.MigrationSlice, error) {
	// set up migration config
	migrator := migrate.NewMigrator(db, &migrations)

	// create / verify tables used to tack migrations
	err := migrator.Init(ctx)

	if err != nil {
		return &migrate.MigrationSlice{}, err
	}

	// grab migration lock to prevent race conditions during migration
	for {
		err := migrator.Lock(ctx)

		if err != nil {
			time.Sleep(1 * time.Second)

			continue
		}

		break
	}

	defer func() {
		unlockErr := migrator.Unlock(ctx)

		if unlockErr != nil {
			logger.Error().Msg(fmt.Sprintf("error %s releasing migration lock after running migrations applied %+v \n unapplied %+v \n last group %+v \n", unlockErr, migrations.Sorted().Applied(), migrations.Sorted().Unapplied(), migrations.Sorted().LastGroup()))
		}
	}()

	// run all un-applied migrations
	group, err := migrator.Migrate(ctx)

	// if migration failed attempt to rollback so migrations can be re-attempted
	if err != nil {
		group, rollbackErr := migrator.Rollback(ctx)

		if rollbackErr != nil {
			return &migrate.MigrationSlice{}, fmt.Errorf("error %s rolling back after original error %s", rollbackErr, err)
		}

		if group.ID == 0 {
			return &migrate.MigrationSlice{}, fmt.Errorf("no groups to rollback after migration error %s", err)
		}

		return &migrate.MigrationSlice{}, fmt.Errorf("rolled back after migration error %s", err)
	}

	// get the status of all run and un-run migrations
	ms, err := migrator.MigrationsWithStatus(ctx)

	if err != nil {
		return &migrate.MigrationSlice{}, err
	}

	if group.ID == 0 {
		logger.Debug().Msg("no new migrations to run")
	}

	return &ms, nil
}

// AwaitDatabaseOnline waits (infinitely) until the provided
// database is reachable
func AwaitDatabaseOnline(serviceDatabase PostgresClient, logger logging.ServiceLogger) {
	var databaseOnline bool

	for !databaseOnline {
		err := serviceDatabase.HealthCheck()

		if err != nil {
			logger.Debug().Msg("unable to connect to database, will retry in 1 second")

			time.Sleep(1 * time.Second)

			continue
		}

		logger.Debug().Msg("connected to database")

		databaseOnline = true
	}
}

const (
	DEFAULT_SCHEMA_NAME = "public"
)

// AwaitTableExists waits (infinitely) until the specified
// table is created
func AwaitTableExists(ctx context.Context, tableName string, schemaName string, serviceDatabase PostgresClient, logger logging.ServiceLogger) {
	var exists bool

	for !exists {
		_ = serviceDatabase.NewRaw(
			"SELECT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = ? AND tablename = ?);", schemaName, tableName,
		).Scan(ctx, &exists)

		logger.Debug().Msgf("table %s not present for schema %s, will retry in 1 second", tableName, schemaName)

		time.Sleep(1 * time.Second)
	}

	logger.Debug().Msgf("table %s present for schema %s, will retry in 1 second", tableName, schemaName)
}
