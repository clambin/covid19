package db

import (
	"embed"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	// postgres sql driver
	_ "github.com/lib/pq"
)

// DB hold the handle to a database.  Provides a Prometheus DBStatsCollector to monitor DB connections
type DB struct {
	Handle   *sqlx.DB
	psqlInfo string
	database string
}

// New creates a new DB connector
func New(cfg configuration.PostgresDB) (db *DB, err error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
	db = &DB{
		psqlInfo: psqlInfo,
		database: cfg.Database,
	}
	if db.Handle, err = sqlx.Connect("postgres", db.psqlInfo); err != nil {
		return
	}

	if err = db.migrate(); err == nil {
		err = prometheus.Register(collectors.NewDBStatsCollector(db.Handle.DB, db.database))
	}
	return
}

func (db *DB) migrate() error {
	migration, err := db.prepareMigration()
	if err == nil {
		err = migration.Up()
	}
	if err == migrate.ErrNoChange {
		err = nil
	}

	return err
}

// RemoveAll deletes all database tables
func (db *DB) RemoveAll() error {
	migration, err := db.prepareMigration()
	if err == nil {
		err = migration.Down()
	}
	return err
}

//go:embed migrations/*
var migrations embed.FS

func (db *DB) prepareMigration() (m *migrate.Migrate, err error) {
	var src source.Driver
	var dbDriver database.Driver

	src, err = iofs.New(migrations, "migrations")
	if err == nil {
		dbDriver, err = postgres.WithInstance(db.Handle.DB, &postgres.Config{DatabaseName: db.database})
	}
	if err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}

	return migrate.NewWithInstance("migrations", src, db.database, dbDriver)
}
