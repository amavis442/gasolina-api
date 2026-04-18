-- ABOUTME: Initial schema for the gasolina database.
-- ABOUTME: deleted_at is nullable for soft-delete sync tracking.

CREATE TABLE IF NOT EXISTS fuel_entries (
    id          TEXT PRIMARY KEY,
    liters      DOUBLE PRECISION NOT NULL,
    total_cost  DOUBLE PRECISION NOT NULL,
    price_per_l DOUBLE PRECISION NOT NULL,
    odometer    DOUBLE PRECISION NOT NULL,
    fuelled_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fuel_entries_updated_at ON fuel_entries (updated_at);
