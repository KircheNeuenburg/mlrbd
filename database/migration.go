package database

import (
	"database/sql"
	"log"
	"strconv"
)

const schemaVersion = 1

// Migrate executes database migrations.
func Migrate(db *sql.DB) {
	var currentVersion int
	db.QueryRow(`select version from schema_version`).Scan(&currentVersion)

	log.Println("Current schema version:", currentVersion)
	log.Println("Latest schema version:", schemaVersion)

	for version := currentVersion + 1; version <= schemaVersion; version++ {
		log.Println("Migrating to version:", version)

		tx, err := db.Begin()
		if err != nil {
			log.Fatal("[Migrate] %v", err)
		}

		rawSQL := SqlMap["schema_version_"+strconv.Itoa(version)]
		_, err = tx.Exec(rawSQL)
		if err != nil {
			tx.Rollback()
			log.Fatal("[Migrate] %v", err)
		}

		if _, err := tx.Exec(`delete from schema_version`); err != nil {
			tx.Rollback()
			log.Fatal("[Migrate] %v", err)
		}

		if _, err := tx.Exec(`insert into schema_version (version) values($1)`, version); err != nil {
			tx.Rollback()
			log.Fatal("[Migrate] %v", err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatal("[Migrate] %v", err)
		}
	}
}
