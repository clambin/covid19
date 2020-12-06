package population

import (
	"fmt"
	"strings"
	"strconv"
	"database/sql"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"

)

// DB interface representing a Population database table
type PopulationDB interface {
	List() (map[string]int64, error)
	Add(map[string]int64) (error)
}

// PostgresDB implements DB in Postgres
type PostgresPopulationDB struct {
	psqlInfo    string
	initialized bool
}

// NewPostgresDB creates a new PostgresDB object
func NewPostgresPopulationDB(host string, port int, database string, user string, password string) (*PostgresPopulationDB) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)

	return &PostgresPopulationDB{psqlInfo: psqlInfo, initialized: false}
}

// List all records from the population table
func (db *PostgresPopulationDB) List() (map[string]int64, error) {
	entries := make(map[string]int64, 0)

	if err := db.initializeDB(); err != nil {
		return entries, err
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err == nil {
		defer dbh.Close()

		rows, err := dbh.Query(fmt.Sprintf( "SELECT country_code, population FROM population"))

		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var code string
				var population int64
				err = rows.Scan(&code, &population)
				if err != nil { break }
				entries[code] = population
			}
			log.Debugf("Found %d records", len(entries))
		}
	}

	return entries, err
}

// Replace replaces searchpattern instances by $<n> in a SQL statement
func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

// Add all specified records in the population database table
func (db *PostgresPopulationDB) Add(entries map[string]int64) (error) {
	if err := db.initializeDB(); err != nil {
		return err
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err != nil { return err }
	defer dbh.Close()

	// Prepare the SQL statement
	sqlStr := "INSERT INTO population(country_code, population) VALUES "
	vals := []interface{}{}

	for code, population := range entries {
		sqlStr += "(?, ?),"
		vals = append(vals, code, population)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	sqlStr = ReplaceSQL(sqlStr, "?")
	sqlStr += "ON CONFLICT (country_code) DO UPDATE SET population = EXCLUDED.population"

	stmt, _ := dbh.Prepare(sqlStr)

	_, err = stmt.Exec(vals...)

	return err
}

// initializeDB creates the required tables
func (db *PostgresPopulationDB) initializeDB() (error) {
	if db.initialized {
		return nil
	}

	dbh, err := sql.Open("postgres", db.psqlInfo)

	if err != nil { return err }
	defer dbh.Close()

	_, err = dbh.Exec(`
		CREATE TABLE IF NOT EXISTS population (
			country_code TEXT PRIMARY KEY,
			population NUMERIC
		)
	`)

	if err == nil {
		db.initialized = true
	}

	return err
}
