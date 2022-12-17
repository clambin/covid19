package db

import (
	"embed"
	"errors"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	// postgres sql driver
	_ "github.com/lib/pq"
)

// DB hold the handle to a database.  Provides a Prometheus DBStatsCollector to monitor DB connections
type DB struct {
	Handle    *sqlx.DB
	Collector prometheus.Collector
	database  string
}

// New creates a new DB connector
func New(cfg configuration.PostgresDB) (*DB, error) {
	dbh, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	db := &DB{
		Handle:    dbh,
		database:  cfg.Database,
		Collector: collectors.NewDBStatsCollector(dbh.DB, cfg.Database),
	}

	if err = db.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, err
}

func (db *DB) migrate() error {
	migration, err := db.prepareMigration()
	if err != nil {
		return fmt.Errorf("prepare migration: %w", err)
	}
	if err = migration.Up(); errors.Is(err, migrate.ErrNoChange) {
		err = nil
	}

	return err
}

// RemoveAll deletes all database tables
func (db *DB) RemoveAll() error {
	migration, err := db.prepareMigration()
	if err != nil {
		return fmt.Errorf("prepare migration: %w", err)
	}
	return migration.Down()
}

//go:embed migrations/*
var migrations embed.FS

func (db *DB) prepareMigration() (*migrate.Migrate, error) {
	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("iofs: %w", err)
	}

	var dbDriver database.Driver
	dbDriver, err = postgres.WithInstance(db.Handle.DB, &postgres.Config{DatabaseName: db.database})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return migrate.NewWithInstance("migrations", src, db.database, dbDriver)
}
