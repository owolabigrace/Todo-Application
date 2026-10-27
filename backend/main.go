package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"todo-app/internal/httpapi"
	"todo-app/internal/store"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	todoStore := store.NewPostgresStore(db)
	server := httpapi.NewServer(todoStore)

	log.Println("Todo API running at http://localhost:8080")

	log.Fatal(server.ListenAndServe())
}