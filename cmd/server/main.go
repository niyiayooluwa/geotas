package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/handler"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to Neon
	var dbURL string = os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	// confirm connection is alive
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	fmt.Println("✅ Connected to Neon successfully")

	// create queries object — this is what talks to the DB
	var queries *db.Queries = db.New(conn)

	// pass queries into the router so handlers can use it
	var router = handler.NewRouter(queries)

	// start server
	var port string = os.Getenv("PORT")
	fmt.Printf("🚀 GEOTAS server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
