# GEOTAS
### Geo-Temporal Attendance System

A university attendance system combining rotating QR codes, geofencing, OTP fallback, and confidence scoring to produce verifiable, tamper-resistant attendance records.

---

## Overview

GEOTAS addresses the core problem of proxy attendance in Nigerian universities. Existing systems rely on paper registers or simple static QR codes — both trivially bypassed by taking photos or signing for friends. 

GEOTAS layers multiple verification mechanisms simultaneously, computing a **confidence score** for every attendance record that reflects how trustworthy that mark is. The novel contribution of this system is the combination of all four mechanisms — rotating QR, geofencing, OTP fallback, and confidence scoring — in a single cohesive platform.

---

## System Architecture

```
Flutter Web (Lecturer Dashboard)
        │
        ├── HTTP: Manage courses, schedules, sessions, & view master attendance
        └── Polling: Fetch real-time rotating QR codes and live attendance updates
                │
         Go Backend (Chi + sqlc + pgx)
                │
         Neon PostgreSQL (Serverless DB)
                │
Flutter Mobile (Student App)
        │
        ├── HTTP: Mark attendance (QR or OTP), join courses, view notifications
        └── Security: GPS coordinates + Device hardware fingerprint sent with every request
```

---

## Core Features & Workflows

### The Lecturer Flow
1. **Registration & Auth**: Signs up on the web dashboard (email strictly checked for `.edu.ng`).
2. **Course Creation**: Creates a course and receives a unique 5-character invite code.
3. **Course Management**: Can invite students, add co-lecturers, set the strictness of the geofence via the `confidence_threshold`, and manage the weekly **Timetable / Schedules**.
4. **Session Management**: Starts an attendance session with a defined geofence radius.
5. **Live Verification**: A rotating QR code is displayed on the projector, refreshing every 30 seconds to prevent screenshot sharing.
6. **Master Records**: The lecturer can export a detailed register showing exact confidence scores and risk factors (like distance or device switching) for every student.

### The Student Flow
1. **SSO Auth**: Logs into the mobile app using Google OAuth (Firebase Auth).
2. **Onboarding**: Joins courses via the lecturer's invite code (requires Matriculation Number).
3. **Real-Time Alerts**: Receives instant in-app **Notifications** whenever a class schedule is updated or a live session begins.
4. **Marking Attendance**: Opens the app inside the classroom (GPS is captured natively).
5. **Scanning**: Scans the rotating QR code on the projector.
6. **Fallback**: If the camera is broken, requests an OTP from the lecturer to mark attendance.

### Co-Lecturer Superpowers
To handle large university classes taught by multiple professors, any lecturer who joins an existing course using an invite code is instantly assigned the `co_lecturer` flag. 
Co-lecturers have full administrative rights over the course (can create sessions, manage schedules, remove students, and view master attendance). Only the original Course Owner can delete the course itself.

---

## Verification Layers

Every attendance mark passes through all of the following checks before being accepted:

| Check | Method |
|---|---|
| **QR Validity** | HMAC-signed token with a strict 30-second expiry window. |
| **Replay Prevention** | Each QR token is single-use and invalidated in the DB upon successful scan. |
| **Geofencing** | Haversine distance computed server-side against the session's center coordinates. |
| **OTP Validity** | Per-user, per-session, 5-minute TTL, single-use fallback codes. |
| **Mock Location** | Android mock provider flag checked on the native device hardware. |
| **Device Fingerprinting** | Device ID tied to attendance record. Duplicate devices in the same session, or switching devices mid-semester, are mathematically penalized. |

---

## Confidence Scoring System

Rather than building a brittle system that simply accepts or rejects attendance, GEOTAS uses a deduction-based scoring algorithm. Every attempt starts at a perfect **1.0 (100%)** and is penalized based on risk factors:

| Factor | Impact |
|---|---|
| **Outside Geofence** | Massive penalty (-0.50). Usually forces score below threshold. |
| **Borderline Distance** | Slight penalty (-0.15) if outside the strict radius but nearby. |
| **Mark Method** | QR scan = full score, OTP fallback = slight penalty (-0.10). |
| **Lateness** | Late scan = (-0.10), Very late scan = (-0.15). |
| **Mock Location** | Heavy penalty (-0.40) and immediate rejection. |
| **Device Switching** | Changing phones from a previous session = (-0.20). |
| **Duplicate Device** | Two students using the exact same phone in one session = (-0.30). |

If the final computed score falls below the lecturer's custom `confidence_threshold` (default 0.75), the API returns a `400 Bad Request` and flatly rejects the attendance.

---

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.21+ with `go-chi` router |
| **Database** | PostgreSQL (hosted on Neon serverless provider) |
| **Query Layer** | `sqlc` (generates type-safe Go code from raw SQL) + `pgx` driver |
| **Authentication** | JWT (`golang-jwt`) for stateless session management + Firebase Auth for OAuth |
| **Mobile App** | Flutter (Dart) |
| **Web Dashboard**| Flutter Web (Dart) |

---

## Database Schema (9 Tables)

The system relies on a heavily normalized Postgres schema:

1. `users` — Single account type containing profile data and Firebase mappings.
2. `courses` — Owned by lecturers, dictates the confidence thresholds.
3. `course_members` — Join table mapping users to courses with roles (`student` or `lecturer`).
4. `sessions` — Live attendance events locking in the geofence coordinates.
5. `qr_tokens` — Ephemeral, rotating signed tokens.
6. `otp_codes` — Fallback validation codes with TTL.
7. `attendance_records` — Final immutable records storing location, method, device fingerprint, and computed confidence score.
8. `schedules` — Stores the weekly timetable slots for the course.
9. `notifications` — Stores ultra-fast bulk-inserted alerts (e.g., `session_starting`, `schedule_updated`) for the mobile dashboard.

---

## Security Considerations & Limitations

**Primary attack vector: GPS spoofing**
A student can fake their location using mock GPS apps. GEOTAS mitigates this by detecting the mock location provider flag natively on Android and applying a heavy confidence score penalty. This is documented as a known limitation — no existing geofence-based system has perfectly solved GPS spoofing without hardware beacons. The confidence score surfaces suspicious records for human review rather than claiming perfect detection.

**Secondary vectors and mitigations:**
- **QR screenshot sharing:** 30-second rotation window limits exposure.
- **OTP sharing:** OTP is tied to a specific user ID; rejected if submitted by another account.
- **Replay attacks:** QR tokens are single-use; stored and checked server-side.
- **Proxy marking via shared device:** Device fingerprinting flags duplicate device IDs in the same session.

---

## Running Locally

```bash
# Clone the repo
git clone https://github.com/niyiayooluwa/geotas
cd geotas

# Install dependencies
go mod download

# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Copy env file and fill in values
cp .env.example .env

# Generate sqlc code
sqlc generate

# Start the server
go run cmd/server/main.go
```

## Environment Variables

```env
DATABASE_URL=postgres://user:pass@host:5432/geotas?sslmode=disable
PORT=8080
JWT_SECRET=your_long_random_secret_for_stateless_sessions
FIREBASE_CREDENTIALS={"type":"service_account",...}
GOOGLE_CLIENT_ID=your_google_oauth_client_id
```

---

## Author

**Toluwalase Abiola Ayooluwa**  
B.Tech Software Engineering — Federal University of Technology, Akure (FUTA)  
Final Year Project — 2026