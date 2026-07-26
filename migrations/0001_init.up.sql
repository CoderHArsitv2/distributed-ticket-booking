-- Phase 2: initial schema for the distributed ticket-booking engine.
-- Targets PostgreSQL. See README "Database Schema Overview".

CREATE TABLE IF NOT EXISTS events (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT        NOT NULL,
    venue           TEXT        NOT NULL,
    total_seats     INT         NOT NULL CHECK (total_seats >= 0),
    available_seats INT         NOT NULL CHECK (available_seats >= 0),
    sale_starts_at  TIMESTAMPTZ NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'UPCOMING'
                        CHECK (status IN ('UPCOMING','ON_SALE','SOLD_OUT','CANCELLED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS seats (
    id             BIGSERIAL PRIMARY KEY,
    event_id       BIGINT      NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    seat_number    TEXT        NOT NULL,
    section        TEXT        NOT NULL,
    price          BIGINT      NOT NULL CHECK (price >= 0), -- minor units
    status         TEXT        NOT NULL DEFAULT 'AVAILABLE'
                       CHECK (status IN ('AVAILABLE','RESERVED','BOOKED','BLOCKED')),
    reserved_until TIMESTAMPTZ,
    version        BIGINT      NOT NULL DEFAULT 0, -- optimistic locking
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, seat_number)
);

-- Sweeper scans RESERVED seats past their hold window.
CREATE INDEX IF NOT EXISTS idx_seats_status_reserved_until
    ON seats (status, reserved_until);

CREATE TABLE IF NOT EXISTS reservations (
    id         BIGSERIAL PRIMARY KEY,
    seat_id    BIGINT      NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    event_id   BIGINT      NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL,
    status     TEXT        NOT NULL DEFAULT 'HELD'
                   CHECK (status IN ('HELD','CONFIRMED','EXPIRED','RELEASED')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reservations_status_expires_at
    ON reservations (status, expires_at);

CREATE TABLE IF NOT EXISTS bookings (
    id           BIGSERIAL PRIMARY KEY,
    event_id     BIGINT      NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id      BIGINT      NOT NULL,
    reference    TEXT        NOT NULL UNIQUE,
    total_amount BIGINT      NOT NULL CHECK (total_amount >= 0),
    status       TEXT        NOT NULL DEFAULT 'PENDING'
                     CHECK (status IN ('PENDING','CONFIRMED','FAILED','REFUNDED')),
    payment_ref  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS booking_seats (
    id             BIGSERIAL PRIMARY KEY,
    booking_id     BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    seat_id        BIGINT NOT NULL REFERENCES seats(id),
    price_snapshot BIGINT NOT NULL CHECK (price_snapshot >= 0),
    UNIQUE (booking_id, seat_id)
);
