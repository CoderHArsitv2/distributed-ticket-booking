package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// LockStrategy selects which concurrency-control engine the reservation
// service uses. See the README for the trade-offs of each.
type LockStrategy string

const (
	LockPessimistic LockStrategy = "pessimistic" // SELECT ... FOR UPDATE
	LockOptimistic  LockStrategy = "optimistic"  // version column CAS
	LockDistributed LockStrategy = "distributed" // Redis SET NX PX + Lua release
)

// Config holds all runtime configuration, populated from the environment.
type Config struct {
	HTTPPort string

	DatabaseURL     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	RedisURL string

	LockStrategy LockStrategy
	HoldTTL      time.Duration // how long a RESERVED hold survives before sweeping
	SweepEvery   time.Duration // expiry sweeper tick interval
}

// Load reads configuration from environment variables, applying sane defaults.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ticketing?sslmode=disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379/0"),
		LockStrategy:    LockStrategy(getEnv("LOCK_STRATEGY", string(LockPessimistic))),
		HoldTTL:         getEnvDuration("HOLD_TTL", 10*time.Minute),
		SweepEvery:      getEnvDuration("SWEEP_EVERY", 30*time.Second),
	}

	switch cfg.LockStrategy {
	case LockPessimistic, LockOptimistic, LockDistributed:
	default:
		return nil, fmt.Errorf("invalid LOCK_STRATEGY %q", cfg.LockStrategy)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
