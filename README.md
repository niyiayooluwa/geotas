# GEOTAS
### Geo-Temporal Attendance System

A university attendance system combining rotating QR codes, geofencing, OTP fallback, and confidence scoring to produce verifiable, tamper-resistant attendance records.

---

## Overview

GEOTAS addresses the core problem of proxy attendance in Nigerian universities. Existing systems rely on paper registers or simple QR codes — both trivially bypassed. GEOTAS layers multiple verification mechanisms simultaneously, computing a confidence score for every attendance record that reflects how trustworthy that mark is.

The novel contribution of this system is the combination of all four mechanisms — rotating QR, geofencing, OTP fallback, and confidence scoring — in a single cohesive system. No existing literature combines all four.

---

## System Architecture

```
Flutter Web (Lecturer Dashboard)
        │
        ├── HTTP: session management, course management
        └── Polling: live attendance updates
                │
         Go Backend (Chi + sqlc + pgx)
                │
         Neon PostgreSQL
                │
Flutter Mobile (Student App)
        │
        ├── HTTP: mark attendance (QR or OTP)
        └── GPS + Device fingerprint sent with every request
```

---

## How It Works

**Lecturer flow:**
1. Creates an account and logs in via the web dashboard
2. Creates a course — receives a unique invite code to share with students
3. Starts an attendance session — sets geofence radius, week number, and optional title
4. A rotating QR code is generated and displayed, refreshing every 30 seconds
5. Watches real-time attendance as students mark in
6. Closes the session when done
7. Exports the printable register at any time

**Student flow:**
1. Creates an account and logs in via the mobile app
2. Joins a course using the lecturer's invite code
3. Opens the app inside the classroom — GPS is captured
4. Scans the rotating QR code to mark attendance
5. If camera fails, requests an OTP — enters the code instead
6. Receives confirmation of attendance

---

## Verification Layers

Every attendance mark passes through all of the following checks before being accepted:

| Check | Method |
|---|---|
| QR validity | HMAC-signed token with 30-second expiry window |
| Replay prevention | Each QR token is single-use, stored in DB |
| Geofence | Haversine distance computed server-side against session coordinates |
| OTP validity | Per-user, per-session, 5-minute TTL, single-use |
| Mock location detection | Android mock provider flag checked on device |
| Device fingerprinting | Device ID tied to attendance record, duplicate device in same session flagged |

---

## Confidence Score

Every attendance record stores a confidence score between `0.0` and `1.0` computed at mark time.

| Factor | Impact |
|---|---|
| Inside geofence | Base requirement — outside heavily penalises score |
| Distance from geofence center | Closer to center = higher score |
| Mark method | QR = full score, OTP fallback = slight penalty |
| Time of scan | Early in session = higher, very late = lower |
| Mock location detected | Heavy penalty |
| Duplicate device ID in session | Penalises both records |

The score is stored permanently and displayed on the exported register, allowing lecturers to flag suspicious records for review rather than automatically rejecting them.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26 + Chi router |
| Database | Neon PostgreSQL (serverless) |
| Query layer | sqlc — type-safe query generation |
| Auth | JWT (golang-jwt/jwt) + bcrypt |
| Mobile | Flutter |
| Web | Flutter Web |
| Hosting | Railway |

---

## Project Structure

```
geotas/
├── cmd/
│   └── server/
│       └── main.go            # entry point
├── internal/
│   ├── db/                    # sqlc generated code
│   ├── handler/               # HTTP handlers — thin layer
│   ├── middleware/            # auth middleware
│   ├── model/                 # shared request/response structs
│   ├── repository/            # database access layer
│   └── service/               # business logic layer
├── migrations/                # SQL migration files
├── query/                     # raw SQL queries for sqlc
├── sqlc.yaml
├── .env.example
└── go.mod
```

---

## Database Schema

Seven tables covering the full attendance lifecycle:

- `users` — single account type with course-scoped roles
- `courses` — owned by lecturers, joined by students via invite code
- `course_members` — course-scoped role assignment (lecturer or student)
- `sessions` — attendance events with geofence coordinates
- `qr_tokens` — rotating signed tokens with expiry and single-use enforcement
- `otp_codes` — per-user per-session fallback codes with TTL
- `attendance_records` — final records with location, method, device fingerprint, and confidence score

---

## API Endpoints

### Public
| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/register` | Create a new account |
| POST | `/auth/login` | Login and receive JWT |

### Protected (requires Bearer token)
| Method | Endpoint | Description |
|---|---|---|
| GET | `/me` | Get current user |
| POST | `/courses` | Create a course |
| GET | `/courses` | List courses you own |
| POST | `/courses/join` | Join a course via invite code |
| POST | `/sessions` | Start an attendance session |
| PATCH | `/sessions/{id}/close` | Close a session |
| GET | `/sessions/{id}/attendance` | Get attendance for a session |
| POST | `/attendance/qr` | Mark attendance via QR |
| POST | `/attendance/otp/request` | Request an OTP |
| POST | `/attendance/otp/verify` | Verify OTP and mark attendance |

---

## Security Considerations

**Primary attack vector: GPS spoofing**
A student can fake their location using mock GPS apps. GEOTAS mitigates this by detecting the mock location provider flag on Android and applying a heavy confidence score penalty. This is documented as a known limitation — no existing geofence-based system has fully solved GPS spoofing. The confidence score surfaces suspicious records for human review rather than claiming perfect detection.

**Secondary vectors and mitigations:**
- QR screenshot sharing → 30-second rotation window limits exposure
- OTP sharing → OTP is tied to a specific user ID, rejected if submitted by another account
- Replay attacks → QR tokens are single-use, stored and checked server-side
- Proxy marking via shared device → device fingerprinting flags duplicate device IDs in the same session

---

## Running Locally

```bash
# clone the repo
git clone https://github.com/niyiayooluwa/geotas
cd geotas

# install dependencies
go mod download

# install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# copy env file and fill in values
cp .env.example .env

# run migrations in your Neon console
# (files are in /migrations in order)

# generate sqlc code
sqlc generate

# start the server
go run cmd/server/main.go
```

---

## Environment Variables

```env
DATABASE_URL=your_neon_connection_string
PORT=8080
JWT_SECRET=your_long_random_secret
```

---

## Author

Toluwalase Abiola Ayooluwa
B.Tech Software Engineering — Federal University of Technology, Akure (FUTA)
Final Year Project — 2026