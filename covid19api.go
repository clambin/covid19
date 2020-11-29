package main

import (
	"log"
	"covid19api/coviddb"
)

const (
	host     = "192.168.0.11"
	port     = 31000
	dbname   = "covid19"
	user     = "covid"
	password = "its4covid"
)

func main() {
	db, err := coviddb.Connect(host, port, dbname, user, password)

	if err != nil {
		panic("failed to connect to database")
	}

	log.Printf("Successfully connected!")

	rows, err := db.List()

	if err != nil {
		panic(err)
	}

	log.Printf("Found %d entries", len(rows))

	totalRows := coviddb.GetTotalCases(rows)
	totalRows =  coviddb.GetTotalDeltas(totalRows)

	for _, totalRow := range totalRows {
		log.Printf("%s-%d-%d-%d", totalRow.Timestamp, totalRow.Confirmed, totalRow.Recovered, totalRow.Deaths)
	}

	db.Close()
}
