-- ABOUTME: Renames the odometer column to kilometers for consistent terminology.
-- ABOUTME: Applies to existing databases; new installs use 001_initial.sql directly.

ALTER TABLE fuel_entries RENAME COLUMN odometer TO kilometers;
