package database

import (
	"crypto/tls"
	"database/sql"
	"fmt"

	"nexus-api/logging"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// PostgresDatabaseConfig contains values for creating a
// new connection to a postgres database
type PostgresDatabaseConfig struct {
	DatabaseName          string
	DatabaseEndpointURL   string
	DatabaseUsername      string
	DatabasePassword      string
	SSLEnabled            bool
	QueryLoggingEnabled   bool
	RunDatabaseMigrations bool
	Logger                *logging.ServiceLogger
}

// PostgresClient wraps a connection to a postgres database
type PostgresClient struct {
	*bun.DB
}

// NewPostgresClient returns a new connection to the specified
// postgres data and error (if any)
func NewPostgresClient(config PostgresDatabaseConfig) (PostgresClient, error) {
	// configure postgres database connection options
	var pgOptions *pgdriver.Connector

	if config.SSLEnabled {
		pgOptions =
			pgdriver.NewConnector(
				pgdriver.WithAddr(config.DatabaseEndpointURL),
				pgdriver.WithUser(config.DatabaseUsername),
				pgdriver.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
				pgdriver.WithPassword(config.DatabasePassword),
				pgdriver.WithDatabase(config.DatabaseName),
			)
	} else {
		pgOptions = pgdriver.NewConnector(
			pgdriver.WithAddr(config.DatabaseEndpointURL),
			pgdriver.WithUser(config.DatabaseUsername),
			pgdriver.WithInsecure(true),
			pgdriver.WithPassword(config.DatabasePassword),
			pgdriver.WithDatabase(config.DatabaseName),
		)
	}

	config.Logger.Debug().Msg(fmt.Sprintf("creating database client with options %+v %+v", pgOptions.Config(), pgOptions.Config().TLSConfig))

	// connect to the database
	sqldb := sql.OpenDB(pgOptions)

	db := bun.NewDB(sqldb, pgdialect.New())

	// set up logging on database if requested
	if config.QueryLoggingEnabled {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return PostgresClient{
		DB: db,
	}, nil
}

// HealthCheck returns an error if the database can not
// be connected to and queried, nil otherwise
func (pg *PostgresClient) HealthCheck() error {
	return pg.Ping()
}
