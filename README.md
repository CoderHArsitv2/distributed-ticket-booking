# Distributed Ticket Booking System (Go)

A high-concurrency, fault-tolerant distributed ticket reservation engine built in Go. Inspired by real-world system design challenges (handling flash sales with 100k+ concurrent users), this project demonstrates how to prevent race conditions and double-booking using multiple concurrency control strategies: **Pessimistic Locking**, **Optimistic Locking**, and **Redis Distributed Locks (Redlock / Lua Scripts)**.

---

## 📌 Project Goals & Inspiration

When high-demand tickets (e.g., major concerts or sports finals) go on sale, thousands of users try to book the exact same seat at the exact same millisecond. Without rigorous concurrency control:
1. **Race conditions** allow two users to read seat `A1` as `AVAILABLE` simultaneously.
2. Both users proceed to checkout and buy seat `A1` — resulting in **double-booking**.

This project implements a production-grade distributed ticketing engine in Go to demonstrate how different locking mechanisms perform under severe write contention, zero double-booking requirements, and sub-second latency targets.

---

## ⚡ Core Concurrency & Locking Strategies Implemented

| Locking Mechanism | Level | Technical Details | Best Suited For |
| :--- | :--- | :--- | :--- |
| **Pessimistic Locking** | Database (PostgreSQL) | Uses `SELECT ... FOR UPDATE` inside a database transaction to lock target seat rows until transaction commit/rollback. | High consistency guarantees, low-to-medium throughput. |
| **Optimistic Locking** | Database (PostgreSQL) | Uses a `version` column in the `seats` table (`UPDATE seats SET status='RESERVED', version=version+1 WHERE seat_id=? AND version=?`). | Low write contention, non-blocking high-read workloads. |
| **Distributed Locking** | Redis (Cache) | Uses Redis `SET key token NX PX 10000` with atomic Lua scripts for release (or `Redsync` for multi-node Redis clusters). | Distributed multi-instance microservices, flash sales. |

---

## 🏗️ System Architecture & Workflow

```
                  +-----------------------+
                  |     Client Request    |
                  +-----------+-----------+
                              |
                              v
                  +-----------------------+
                  |  API Gateway / Router |
                  +-----------+-----------+
                              |
         +--------------------+--------------------+
         |                                         |
         v                                         v
+------------------+                    +---------------------+
|  Event Service   |                    | Reservation Service |
| (Search/Catalog) |                    |  (Locking Engine)   |
+------------------+                    +----------+----------+
                                                   |
                     +-----------------------------+-----------------------------+
                     |                             |                             |
                     v                             v                             v
           +-------------------+         +-------------------+         +-------------------+
           | Pessimistic Lock  |         |  Optimistic Lock  |         | Distributed Lock  |
           | (SELECT FOR UPD)  |         | (Version Control) |         | (Redis + Lua Script)
           +---------+---------+         +---------+---------+         +---------+---------+
                     |                             |                             |
                     +-----------------------------+-----------------------------+
                                                   |
                                                   v
                                        +--------------------+
                                        |     PostgreSQL     |
                                        +--------------------+
```

### High-Level Booking Flow
1. **Seat Selection & Hold**: User requests a seat hold. The system acquires a lock based on the configured strategy.
2. **Temporary Hold (Reservation)**: If successful, the seat status is set to `RESERVED` with a 10-minute expiration window.
3. **Payment Checkout**: User pays via payment gateway.
4. **Booking Confirmation**: On payment success, status transitions to `BOOKED` and a `booking` record is generated.
5. **Expiry Sweeper**: A background worker releases expired `RESERVED` seats back to `AVAILABLE` state if payment is not completed in time.

---

## 📂 Project Structure

The project uses a layered architecture. Each package has a single responsibility, and dependencies flow in one direction: `controllers → service → {locking, repository} → models`.

```
distributed-ticket-booking/
├── cmd/
│   └── server/
│       └── main.go            # Entrypoint: load config → select lock strategy → wire layers → serve (graceful shutdown)
├── config/
│   └── config.go              # Env-driven configuration (DB URL, Redis URL, lock strategy, hold TTL, sweeper cadence)
├── models/                    # Domain types + status enums (no logic)
│   ├── event.go               #   Event        + EventStatus  (UPCOMING/ON_SALE/SOLD_OUT/CANCELLED)
│   ├── seat.go                #   Seat         + SeatStatus   (AVAILABLE/RESERVED/BOOKED/BLOCKED) — has `version` for optimistic lock
│   ├── reservation.go         #   Reservation  + ReservationStatus (HELD/CONFIRMED/EXPIRED/RELEASED)
│   └── booking.go             #   Booking, BookingSeat + BookingStatus (PENDING/CONFIRMED/FAILED/REFUNDED)
├── repository/
│   └── repository.go          # Repository-pattern interfaces (Event/Seat/Reservation/Booking) — Postgres impls land in Phase 3
├── locking/                   # Concurrency-control engine (the core of the project)
│   ├── locker.go              #   SeatLocker interface + shared errors (ErrSeatUnavailable, ErrLockLost)
│   ├── pessimistic.go         #   Strategy 1: SELECT ... FOR UPDATE
│   ├── optimistic.go          #   Strategy 2: version-column compare-and-swap
│   └── distributed.go         #   Strategy 3: Redis SET NX PX + atomic Lua release script
├── service/                   # Business logic / orchestration
│   ├── reservation.go         #   Hold → reserve flow, delegates to configured SeatLocker
│   └── sweeper.go             #   Background worker: releases expired RESERVED seats back to AVAILABLE
├── controllers/               # HTTP layer (no business logic)
│   ├── response.go            #   JSON write / error helpers
│   ├── health_controller.go   #   GET /healthz
│   └── reservation_controller.go # POST /api/v1/reservations
├── middleware/
│   └── middleware.go          # Request logging, panic recovery, middleware chaining
├── router/
│   └── router.go              # Route table + middleware wiring (stdlib net/http mux)
├── migrations/
│   ├── 0001_init.up.sql       # PostgreSQL schema: events, seats, reservations, bookings, booking_seats
│   └── 0001_init.down.sql     # Rollback
├── .env.example               # Sample configuration
├── Makefile                   # run / build / test / migrate-up / migrate-down targets
├── go.mod
└── README.md
```

---

## 🔄 Code Flow (Request → Response)

A seat-hold request travels through the layers as follows:

```
HTTP POST /api/v1/reservations
        │
        ▼
router ──(middleware: Recover → Logger)──► controllers.ReservationController.Hold
        │  decode JSON, validate input
        ▼
service.ReservationService.HoldSeat(eventID, seatID, userID)
        │  orchestrates the booking flow
        ▼
locking.SeatLocker.Acquire(seatID, userID)     ◄── strategy chosen at startup via LOCK_STRATEGY
        │      ├── pessimistic → BEGIN; SELECT ... FOR UPDATE; UPDATE→RESERVED; COMMIT
        │      ├── optimistic  → UPDATE seats SET status='RESERVED', version=version+1 WHERE id=? AND version=?
        │      └── distributed → SET lock:<seatID> <token> NX PX <ttl>   (release via Lua CAS-delete)
        ▼
repository.SeatRepository / ReservationRepository   ──► PostgreSQL
        │  persist RESERVED seat + reservation row (expires_at = now + HOLD_TTL)
        ▼
controllers ──► 201 Created { status: "HELD", seat_id, strategy }
```

**Background flow (Expiry Sweeper — Phase 5):** `service.Sweeper` ticks every `SWEEP_EVERY`, calling `SeatRepository.ReleaseExpired` + `ReservationRepository.MarkExpired` to return unpaid `RESERVED` seats to `AVAILABLE`.

The active locking strategy is selected once at startup in [`main.go`](cmd/server/main.go) from the `LOCK_STRATEGY` env var (`pessimistic` | `optimistic` | `distributed`), so the same codebase can be benchmarked under each strategy without changes.

---

## 🗄️ Database Schema Overview

- **`events`**: Canonical record for events, venues, total/available seats, sale start time, and status (`UPCOMING`, `ON_SALE`, `SOLD_OUT`, `CANCELLED`).
- **`seats`**: Inventory rows containing `seat_number`, `section`, `price`, `status` (`AVAILABLE`, `RESERVED`, `BOOKED`, `BLOCKED`), `reserved_until`, and `version` (for optimistic locking).
- **`reservations`**: Ephemeral holds tracking `user_id`, `expires_at`, and `status`.
- **`bookings`**: Finalized customer orders with payment details and unique reference code.
- **`booking_seats`**: Junction table preserving purchase price snapshots per seat.

---

## 🧰 Tech Stack & Tools

- **Language**: Go 1.23+
- **HTTP Routing**: Go 1.22+ stdlib `net/http` method-pattern mux (pluggable to Chi/Gin in Phase 6)
- **Database**: **PostgreSQL** (with `pgx`/`sqlx` and raw SQL migrations via `golang-migrate`)
- **Distributed Cache & Locking**: Redis (using `go-redis` and atomic Lua scripts)
- **Concurrency & Load Testing**: Go Goroutines, `golang.org/x/sync/errgroup`, custom CLI benchmark runner

---

## 🚀 Getting Started

**Prerequisites:** Go 1.23+, PostgreSQL 14+, Redis 6+ (only needed for the distributed strategy).

```bash
# 1. Configure
cp .env.example .env          # then edit DATABASE_URL / REDIS_URL / LOCK_STRATEGY

# 2. Create the database and run migrations (requires golang-migrate)
createdb ticketing
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/ticketing?sslmode=disable"
make migrate-up

# 3. Run the server
make run                      # serves on :8080 by default

# 4. Smoke test
curl localhost:8080/healthz
```

Switch the concurrency strategy without touching code by setting `LOCK_STRATEGY` to `pessimistic`, `optimistic`, or `distributed` in `.env`.

---

## 🗺️ Project Roadmap

- [x] **Phase 1: Project Initialization & Architectural Specifications** (README & Design)
- [ ] **Phase 2: Database Schema & Migration Setup**
- [ ] **Phase 3: Core Domain Models & Repository Pattern in Go**
- [ ] **Phase 4: Concurrency Engine Implementation**
  - [ ] DB Pessimistic Locking Strategy
  - [ ] DB Optimistic Locking Strategy
  - [ ] Redis Distributed Lock Strategy (Lua script atomic release)
- [ ] **Phase 5: Reservation Hold Expiry Sweeper Service**
- [ ] **Phase 6: REST API Handlers & HTTP Middleware**
- [ ] **Phase 7: Concurrent Stress Test & Load Benchmarking Suite** (Simulating 100k requests to verify zero double-booking)

---

## 📖 Reference Article

This project is inspired by the system design article:
[*Building a Ticketing System: Concurrency, Locks, and Race Conditions*](https://freedium-mirror.cfd/medium.com/@codefarm0/building-a-ticketing-system-concurrency-locks-and-race-conditions-182e0932d962) by Arvind Kumar.
