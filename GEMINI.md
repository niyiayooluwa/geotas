# Project: GEOTAS (Geo-Temporal Attendance System)

This document provides a comprehensive guide for an AI assistant to understand and work with the GEOTAS codebase.

## 1. Project Overview

GEOTAS is a university attendance system designed to combat proxy attendance. It utilizes a multi-layered verification approach:

*   **Rotating QR Codes:** Time-sensitive QR codes that expire every 30 seconds to prevent screenshot sharing.
*   **Geofencing:** Server-side validation of a student's location using Haversine distance calculation.
*   **OTP Fallback:** A one-time password system for situations where QR scanning is not possible.
*   **Confidence Scoring:** A unique system that assigns a trust score to each attendance record based on various factors (e.g., location accuracy, time of scan, use of OTP).

The system consists of a Go backend, a Flutter web dashboard for lecturers, and a Flutter mobile app for students.

## 2. Tech Stack

*   **Backend:** Go (v1.21+) with the Chi router.
*   **Database:** PostgreSQL (designed for Neon serverless provider).
*   **ORM/Query Layer:** `sqlc` for generating type-safe Go code from raw SQL queries.
*   **Authentication:** JWT (JSON Web Tokens).
*   **Frontend:** Flutter (for both web and mobile).

## 3. Architecture & Project Structure

The backend follows a layered architecture to separate concerns.

```
geotas/
├── cmd/server/main.go   # Application entry point, server setup
├── internal/
│   ├── db/              # sqlc-generated code (DO NOT EDIT MANUALLY)
│   ├── handler/         # HTTP handlers (thin, minimal logic)
│   ├── middleware/      # Request middleware (e.g., auth)
│   ├── model/           # Shared Go structs for requests/responses
│   ├── repository/      # Database access layer, abstracts db operations
│   └── service/         # Core business logic
├── migrations/          # SQL database migration files (sequential)
├── query/               # Raw SQL queries for sqlc to process
├── sqlc.yaml            # Configuration for sqlc code generation
├── go.mod               # Go module definitions
└── README.md            # Detailed project documentation
```

**Workflow:** `main.go` -> `handler` -> `service` -> `repository` -> `db` (sqlc)

## 4. Development Workflow

### Running the Application

1.  **Prerequisites:**
    *   Go (latest version recommended)
    *   `sqlc` installed (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
    *   A running PostgreSQL database.

2.  **Setup:**
    *   Create a `.env` file from `.env.example` and populate it with your database URL and JWT secret.
    *   Run database migrations from the `migrations/` directory against your database. They must be run in numerical order.

3.  **Run:**
    ```bash
    go run cmd/server/main.go
    ```

### Database Changes & `sqlc`

**This is a critical workflow.** The project uses `sqlc` to generate type-safe database access code. You **do not** write database queries directly in Go code.

1.  **Write SQL Queries:** Add or modify `.sql` files in the `query/` directory. Use the required `sqlc` comment annotations (e.g., `-- name: GetUser :one`).
2.  **Generate Go Code:** After changing the queries, run `sqlc generate`. This will update the code in the `internal/db/` directory.
3.  **Use the Generated Code:** The newly generated functions in `internal/db/` can now be used in the `repository` layer.

**If you need to change the database schema:**

1.  Create a new, numbered migration file in the `migrations/` directory (e.g., `010_add_new_column.sql`).
2.  Write the `ALTER TABLE` or `CREATE TABLE` statements in this new file.
3.  Apply the migration to your local database.
4.  Update the relevant SQL queries in the `query/` directory to reflect the schema change.
5.  Run `sqlc generate`.

## 5. Coding Conventions

*   **Error Handling:** All errors are propagated up to the handler layer and handled there. Services and repositories should return errors.
*   **Dependency Injection:** Dependencies (like services and repositories) are explicitly passed into the structs that need them (see `cmd/server/main.go` for assembly).
*   **Configuration:** All configuration is managed via environment variables, loaded at startup.
*   **Logging:** Use the standard `log` package for now.
*   **API:** The API is RESTful and uses JSON for serialization. See the `README.md` for a full list of endpoints.
