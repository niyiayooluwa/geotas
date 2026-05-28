package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/handler"
	"github.com/niyiayooluwa/geotas/internal/repository"
	"github.com/niyiayooluwa/geotas/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found. Reading credentials from system environment variables.")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Fatal Error: DATABASE_URL environment variable is completely empty or missing.")
	}

	conn, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close()

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	fmt.Println("✅ Connected to Neon successfully")

	firebaseCreds := os.Getenv("FIREBASE_CREDENTIALS")
	if firebaseCreds == "" {
		log.Fatal("Fatal Error: FIREBASE_CREDENTIALS environment variable is missing.")
	}

	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsJSON([]byte(firebaseCreds)))
	if err != nil {
		log.Fatalf("Error initialising Firebase app: %v\n", err)
	}

	firebaseClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v\n", err)
	}
	fmt.Println("✅ Firebase Admin SDK initialised")

	queries := db.New(conn)
	userRepo := repository.NewUserRepository(queries)
	authService := service.NewAuthService(userRepo, firebaseClient)

	var router = handler.NewRouter(queries, authService)

	var port string = os.Getenv("PORT")
	fmt.Printf("🚀 GEOTAS server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}