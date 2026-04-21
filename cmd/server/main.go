package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/niyiayooluwa/geotas/internal/handler"
)

func main() {
	// Load .env file into the environment
	// If it fails, kill the program — no point running without config
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Read the database URL from .env
	var dbURL string = os.Getenv("DATABASE_URL")

	// Open a connection to Neon using that URL
	// pgx.Connect returns two things — the connection and an error
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// defer means "run this when main() exits"
	// ensures the DB connection is always closed cleanly
	defer conn.Close(context.Background())

	// Ping confirms the connection is actually alive
	// Connect succeeding doesn't guarantee the DB is reachable
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	fmt.Println("✅ Connected to Neon successfully")

	var router = handler.NewRouter()

	// Read port from .env
	var port string = os.Getenv("PORT")
	fmt.Printf("🚀 GEOTAS server running on port %s\n", port)

	// Start the HTTP server on the given port
	// pass the router so it handles all incoming requests
	// previously we passed nil here — that's why /health returned 404
	// nil meant no router, no routes, nothing to handle requests
	log.Fatal(http.ListenAndServe(":"+port, router))
}
