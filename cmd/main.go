package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper/internal/endpoints"
	"github.com/spf13/viper/internal/handlers"
	"log"
)

var InitDB = `
CREATE TABLE IF NOT EXISTS products (
	id VARCHAR(36) PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	description VARCHAR(200),
	price INTEGER default 0,
	quantity INTEGER default 0);`

func main() {

	db, err := sql.Open("sqlite3", "market.db")

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	fmt.Println("Creating database")
	_, err = db.Exec(InitDB)
	if err != nil {
		log.Fatal(err)
	}

	products := handlers.NewProductRoutes(db)
	r := gin.Default()

	endpoints.InitEndpoints(r, products)
	err = r.Run(":8080")

	if err != nil {
		log.Fatal(err)
		return
	}
}
