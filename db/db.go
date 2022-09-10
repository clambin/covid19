package db

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/clambin/covid19/configuration"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	// postgres sql driver
	_ "github.com/lib/pq"
)

// DB hold the handle to a database.  Provides a Prometheus DBStatsCollector to monitor DB connections
type DB struct {
	Handle   *sql.DB
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
	if db.Handle, err = sql.Open("postgres", db.psqlInfo); err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = db.Handle.Close()
		}
	}()

	if err = db.Handle.Ping(); err != nil {
		return
	}
	if err = db.migrate(); err != nil {
		return
	}
	err = prometheus.Register(collectors.NewDBStatsCollector(db.Handle, db.database))
	return
}

func (db *DB) migrate() error {
	m, err := db.prepareMigration()
	if err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	if err = m.Up(); err == migrate.ErrNoChange {
		err = nil
	}

	return err
}

func (db *DB) RemoveAll() error {
	m, err := db.prepareMigration()
	if err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	return m.Down()
}

//go:embed migrations/*
var migrations embed.FS

func (db *DB) prepareMigration() (m *migrate.Migrate, err error) {
	var src source.Driver
	if src, err = iofs.New(migrations, "migrations"); err != nil {
		return nil, fmt.Errorf("invalid migration source: %w", err)
	}

	var dbDriver database.Driver
	if dbDriver, err = postgres.WithInstance(db.Handle, &postgres.Config{DatabaseName: db.database}); err != nil {
		return nil, fmt.Errorf("invalid migration target: %w", err)
	}

	return migrate.NewWithInstance("migrations", src, db.database, dbDriver)
}
