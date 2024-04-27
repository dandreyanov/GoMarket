package main

import (
	"GoMarket/internal/endpoints"
	"GoMarket/internal/handlers"
	"GoMarket/internal/storage"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {

	db, err := storage.Storage()

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	products := handlers.NewProductRoutes(db)
	r := gin.Default()

	endpoints.InitEndpoints(r, products)
	err = r.Run(":8080")

	if err != nil {
		log.Fatal(err)
		return
	}
}
