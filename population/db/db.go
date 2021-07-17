package db

import (
	"database/sql"
	"fmt"
	// postgres sql driver
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// DB interface representing a Population database table
type DB interface {
	List() (map[string]int64, error)
	Add(string, int64) error
}

// PostgresDB implements DB in Postgres
type PostgresDB struct {
	psqlInfo string
	database string
	dbh      *sql.DB
}

// NewPostgresDB creates a new PostgresDB object
func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	return &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		database: database,
	}
}

// List all records from the population table
func (db *PostgresDB) List() (entries map[string]int64, err error) {
	db.initializeDB()

	var rows *sql.Rows
	rows, err = db.dbh.Query(fmt.Sprintf("SELECT country_code, population FROM population"))

	if err != nil {
		db.close()
		return nil, fmt.Errorf("failed to get population data from database: %v", err)
	}

	entries = make(map[string]int64, 0)
	for rows.Next() {
		var code string
		var population int64
		if rows.Scan(&code, &population) == nil {
			entries[code] = population
		}
	}
	_ = rows.Close()

	return entries, err
}

// Add to population database table. If a record for the specified country code already exists, it will be updated
func (db *PostgresDB) Add(code string, population int64) (err error) {
	db.initializeDB()

	sqlStr := fmt.Sprintf("INSERT INTO population(country_code, population) VALUES ('%s', %d) "+
		"ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population",
		code, population)

	var stmt *sql.Stmt
	stmt, err = db.dbh.Prepare(sqlStr)
	if err == nil {
		_, err = stmt.Exec()
	}

	if err != nil {
		err = fmt.Errorf("failed to insert population data in database: %v", err)
		db.close()
	}

	return
}

// initializeDB creates the required tables
func (db *PostgresDB) initializeDB() {
	if db.dbh != nil {
		return
	}

	var err error
	db.dbh, err = sql.Open("postgres", db.psqlInfo)

	if err != nil {
		log.WithError(err).Fatalf("failed to open database '%s'", db.database)
	}

	// r := prometheus.NewRegistry()
	// r.MustRegister(collectors.NewDBStatsCollector(db.dbh, db.database))

	_, err = db.dbh.Exec(`
		CREATE TABLE IF NOT EXISTS population (
			country_code TEXT PRIMARY KEY,
			population NUMERIC
		)
	`)

	if err != nil {
		log.WithError(err).Fatalf("failed to initialize database '%s'", db.database)
	}

	return
}

func (db *PostgresDB) close() {
	if db.dbh != nil {
		_ = db.dbh.Close()
	}
	db.dbh = nil
}
