// ABOUTME: PostgreSQL implementation of the FuelEntry repository interface.
// ABOUTME: Uses pgx/v5; SQL and DB details must not leak into other layers.

package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

type FuelEntryRepository struct {
	db *pgxpool.Pool
}

func NewFuelEntryRepository(db *pgxpool.Pool) *FuelEntryRepository {
	return &FuelEntryRepository{db: db}
}

func (r *FuelEntryRepository) GetAll(ctx context.Context) ([]*fuelentry.FuelEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, liters, total_cost, price_per_l, kilometers, fuelled_at, created_at, updated_at, deleted_at
		FROM fuel_entries ORDER BY fuelled_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *FuelEntryRepository) GetSince(ctx context.Context, since time.Time) ([]*fuelentry.FuelEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, liters, total_cost, price_per_l, kilometers, fuelled_at, created_at, updated_at, deleted_at
		FROM fuel_entries WHERE updated_at > $1 ORDER BY fuelled_at DESC`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func (r *FuelEntryRepository) GetByID(ctx context.Context, id string) (*fuelentry.FuelEntry, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, liters, total_cost, price_per_l, kilometers, fuelled_at, created_at, updated_at, deleted_at
		FROM fuel_entries WHERE id = $1`, id)
	return scanRow(row)
}

func (r *FuelEntryRepository) Save(ctx context.Context, e *fuelentry.FuelEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO fuel_entries (id, liters, total_cost, price_per_l, kilometers, fuelled_at, created_at, updated_at, deleted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.Liters, e.TotalCost, e.PricePerL, e.Kilometers, e.FuelledAt, e.CreatedAt, e.UpdatedAt, e.DeletedAt)
	return err
}

func (r *FuelEntryRepository) Update(ctx context.Context, e *fuelentry.FuelEntry) error {
	_, err := r.db.Exec(ctx, `
		UPDATE fuel_entries SET liters=$2, total_cost=$3, price_per_l=$4, kilometers=$5,
		fuelled_at=$6, updated_at=$7, deleted_at=$8 WHERE id=$1`,
		e.ID, e.Liters, e.TotalCost, e.PricePerL, e.Kilometers, e.FuelledAt, e.UpdatedAt, e.DeletedAt)
	return err
}

func (r *FuelEntryRepository) Delete(ctx context.Context, id string, deletedAt time.Time) error {
	ct, err := r.db.Exec(ctx, `
		UPDATE fuel_entries SET deleted_at=$2, updated_at=$2 WHERE id=$1`, id, deletedAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("entry not found")
	}
	return nil
}
