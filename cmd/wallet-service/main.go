package main

import (
	"log"
	"net/http"
	"os"

	"golang-wallet/internal/db"
	"golang-wallet/internal/handler"
)

func main() {
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Could not connect to the database: %s", err)
	}
	defer database.Close()

    err = db.ApplyMigration(database, "migrations/001_create_wallets_table.sql")
    if err != nil {
        log.Fatalf("Error applying migration: %v", err)
    }


	http.HandleFunc("/api/v1/wallet", handler.WalletOperationHandler(database))
	http.HandleFunc("/api/v1/wallets/", handler.GetBalanceHandler(database))


	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}

}