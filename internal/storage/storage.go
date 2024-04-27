package storage

import (
	"database/sql"
	"log"
)

var InitDB = `
CREATE TABLE IF NOT EXISTS products (
	id VARCHAR(36) PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	description VARCHAR(200),
	price INTEGER default 0,
	quantity INTEGER default 0);`

func Storage() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "market.db")

	log.Println("Creating database")

	_, err = db.Exec(InitDB)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}
