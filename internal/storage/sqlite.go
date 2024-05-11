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
	quantity INTEGER default 0);
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100));
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    quantity INTEGER NOT NULL,
    price INTEGER NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (product_id) REFERENCES products(id));`

func Storage() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "market.db")

	log.Println("Creating database")

	_, err = db.Exec(InitDB)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}
