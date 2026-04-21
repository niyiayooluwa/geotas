package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and returns a configured Chi router
// with all routes registered
func NewRouter() *chi.Mux {
	// chi.NewRouter() creates a new router
	// think of it as the traffic director —
	// it decides which function handles which URL
	var router *chi.Mux = chi.NewRouter()

	// middleware runs on every request before it hits your handler
	// middleware.Logger prints every incoming request to the terminal
	// so you can see GET /health, POST /auth/login etc in real time
	router.Use(middleware.Logger)

	// middleware.Recoverer catches any panic in your handlers
	// instead of crashing the whole server, it returns a 500 error
	// and keeps the server running
	router.Use(middleware.Recoverer)

	// This registers a GET endpoint at /health
	// when someone hits GET http://localhost:8080/health
	// this function runs
	// w is the response writer — you use it to send back a response
	// r is the incoming request — you use it to read headers, body, params
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// 200 OK status code
		w.WriteHeader(http.StatusOK)
		// write the response body
		w.Write([]byte("GEOTAS is alive"))
	})

	return router
}
