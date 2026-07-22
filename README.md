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
| **Pessimistic Locking** | Database (SQL) | Uses `SELECT ... FOR UPDATE` inside a database transaction to lock target seat rows until transaction commit/rollback. | High consistency guarantees, low-to-medium throughput. |
| **Optimistic Locking** | Database (SQL) | Uses a `version` column in the `seats` table (`UPDATE seats SET status='RESERVED', version=version+1 WHERE seat_id=? AND version=?`). | Low write contention, non-blocking high-read workloads. |
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
                                        | Postgres / MySQL   |
                                        +--------------------+
```

### High-Level Booking Flow
1. **Seat Selection & Hold**: User requests a seat hold. The system acquires a lock based on the configured strategy.
2. **Temporary Hold (Reservation)**: If successful, the seat status is set to `RESERVED` with a 10-minute expiration window.
3. **Payment Checkout**: User pays via payment gateway.
4. **Booking Confirmation**: On payment success, status transitions to `BOOKED` and a `booking` record is generated.
5. **Expiry Sweeper**: A background worker releases expired `RESERVED` seats back to `AVAILABLE` state if payment is not completed in time.

---

## 🗄️ Database Schema Overview

- **`events`**: Canonical record for events, venues, total/available seats, sale start time, and status (`UPCOMING`, `ON_SALE`, `SOLD_OUT`, `CANCELLED`).
- **`seats`**: Inventory rows containing `seat_number`, `section`, `price`, `status` (`AVAILABLE`, `RESERVED`, `BOOKED`, `BLOCKED`), `reserved_until`, and `version` (for optimistic locking).
- **`reservations`**: Ephemeral holds tracking `user_id`, `expires_at`, and `status`.
- **`bookings`**: Finalized customer orders with payment details and unique reference code.
- **`booking_seats`**: Junction table preserving purchase price snapshots per seat.

---

## 🧰 Tech Stack & Tools

- **Language**: Go 1.22+
- **HTTP Routing**: Chi / Gin
- **Database**: PostgreSQL / MySQL with `sqlx` / raw SQL migrations
- **Distributed Cache & Locking**: Redis (using `go-redis` and atomic Lua scripts)
- **Concurrency & Load Testing**: Go Goroutines, `golang.org/x/sync/errgroup`, custom CLI benchmark runner

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
