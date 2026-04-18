-- ABOUTME: Creates the application database user and grants table privileges.
-- ABOUTME: Run as a PostgreSQL superuser (e.g. postgres) after 001_initial.sql.

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'gasolina') THEN
        CREATE USER gasolina WITH PASSWORD 'gasolina';
    END IF;
END
$$;

GRANT CONNECT ON DATABASE gasolina TO gasolina;
GRANT USAGE ON SCHEMA public TO gasolina;
GRANT ALL PRIVILEGES ON TABLE fuel_entries TO gasolina;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gasolina;

-- Ensure future tables created by the superuser are also accessible
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL PRIVILEGES ON TABLES TO gasolina;
