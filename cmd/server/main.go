package main

import (
	// context is used for carrying request-scoped data, cancelation signals, and deadlines across API boundaries.
	"context"
	// fmt provides formatted input/output functions, like printing to the console.
	"fmt"
	// log provides a simple logging package for printing error messages and stopping the program if a fatal error occurs.
	"log"
	// net/http provides HTTP client and server implementations, allowing this app to run as a web server.
	"net/http"
	// os provides a platform-independent interface to operating system functionality, like reading environment variables.
	"os"

	// pgxpool is a connection pool for PostgreSQL, allowing the app to efficiently manage multiple database connections.
	"github.com/jackc/pgx/v5/pgxpool"
	// godotenv is a library used to load environment variables from a .env file into the application's environment.
	"github.com/joho/godotenv"

	// db contains the auto-generated code for interacting with the database.
	"github.com/niyiayooluwa/geotas/internal/db"
	// handler contains the code that handles incoming HTTP requests (the "controllers").
	"github.com/niyiayooluwa/geotas/internal/handler"
)

// main is the primary function that runs when the application is started.
func main() {
	// 1. Load Configuration
	// We read settings from a '.env' file. This file usually contains secrets and configurations
	// like the database password and the port the server should run on, which shouldn't be in the code directly.
	//  SAFE FOR PRODUCTION
	// godotenv.Load() returns an error if the file doesn't exist.
	// We log it as an info message, NOT a fatal crash.
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found. Reading credentials from system environment variables.")
	}

	// 2. Connect to the Database
	// We retrieve the 'DATABASE_URL' environment variable, which tells us how to connect to our PostgreSQL database (Neon).
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Fatal Error: DATABASE_URL environment variable is completely empty or missing.")
	}

	// Create a new connection pool using the provided URL. A connection pool manages multiple simultaneous database connections.
	conn, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		// If we can't create the connection pool (e.g., wrong password, database is down), stop the program.
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// 'defer' ensures that the connection pool is properly closed when the main function finishes running,
	// even if an error occurs later on. This prevents memory leaks.
	defer conn.Close()

	// 3. Verify Database Connection
	// Just creating the pool doesn't guarantee a successful connection. 'Ping' actively checks if the database is reachable.
	if err := conn.Ping(context.Background()); err != nil {
		// If the ping fails, log an error and stop.
		log.Fatalf("Database ping failed: %v\n", err)
	}
	// Print a success message to the console if the connection is established.
	fmt.Println("✅ Connected to Neon successfully")

	// 4. Set Up Database Queries Layer
	// We initialize our auto-generated database query layer, passing it the connection pool we just created.
	// This 'queries' object provides type-safe functions to interact with the database (like fetching users, adding attendance).
	var queries *db.Queries = db.New(conn)

	// 5. Set Up the Router
	// The router is responsible for taking an incoming web request (e.g., a GET request to "/users")
	// and directing it to the correct piece of code (handler) to process it.
	// We pass the 'queries' object to the router so that the handlers can access the database.
	var router = handler.NewRouter(queries)

	// 6. Start the Web Server
	// We retrieve the 'PORT' environment variable, which defines which port the server should listen on (e.g., 8080).
	var port string = os.Getenv("PORT")
	// Print a message indicating the server is starting.
	fmt.Printf("🚀 GEOTAS server running on port %s\n", port)

	// http.ListenAndServe starts the web server on the specified port, using our configured router.
	// This function runs continuously and only returns an error if the server crashes.
	// We wrap it in log.Fatal so that if it does crash, the error is logged and the program exits.
	log.Fatal(http.ListenAndServe(":"+port, router))
}