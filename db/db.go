package db

import (
	"database/sql"
	"log"
	// Postgresql driver import
	_ "github.com/lib/pq"
)

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, minConnections, maxConnections int) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)

	return db, nil
}

func GetDbGroups(db *sql.DB) (dg []string, err error) {
	sel := "SELECT ldap_group FROM group_mapping"
	var lg_rows *sql.Rows
	lg_rows, err = db.Query(sel)
	if err != nil {
		log.Fatal(err)
	}
	for lg_rows.Next() {
		var lg string
		if err := lg_rows.Scan(&lg); err != nil {
			log.Fatal(err)
		}
		dg = append(dg, lg)
	}
	return
}

func GetDbRooms(db *sql.DB) (dr []string, err error) {
	sel := "SELECT matrix_room FROM group_mapping"
	var mr_rows *sql.Rows
	mr_rows, err = db.Query(sel)
	if err != nil {
		log.Fatal(err)
	}
	for mr_rows.Next() {
		var mr string
		if err := mr_rows.Scan(&mr); err != nil {
			log.Fatal(err)
		}
		dr = append(dr, mr)
	}
	return
}
