package crt

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	connString = "user=guest dbname=certwatch sslmode=verify-full host=crt.sh port=5432"
)

type config struct {
	db     *sql.DB
	dbOpen bool
}

var (
	conf config
)

func OpenDatabaseConnection() error {
	var err error
	conf.db, err = sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("could not connect to crt.sh DB: %v", err)
	}

	conf.dbOpen = true

	return nil
}

// QueryDomain finds all unique subdomains ever used with a given domain
func QueryDomain(domain string) (*[]string, error) {
	if !conf.dbOpen {
		return nil, fmt.Errorf("open the database before use")
	}

	rows, err := conf.db.Query("SELECT DISTINCT NAME_VALUE FROM certificate_identity ci WHERE reverse(lower(ci.NAME_VALUE)) LIKE reverse(lower($1))", fmt.Sprintf("%%.%s", domain))
	if err != nil {
		return nil, fmt.Errorf("error querying database: %v", err)
	}

	subDomains := make([]string, 1)
	subDomains[0] = domain
	for rows.Next() {
		t := ""
		subDomains = append(subDomains, t)
		err := rows.Scan(&subDomains[len(subDomains)-1])
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %v", err)
		}
	}

	return &subDomains, nil
}

// Dispose makes sure to close the connection to the database
func Dispose() error {
	conf.dbOpen = false
	return conf.db.Close()
}
