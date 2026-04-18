// ABOUTME: Row scanning helpers for pgx query results into FuelEntry entities.
// ABOUTME: Kept separate to avoid repeating scan logic across repository methods.

package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

func scanRow(row pgx.Row) (*fuelentry.FuelEntry, error) {
	var e fuelentry.FuelEntry
	err := row.Scan(&e.ID, &e.Liters, &e.TotalCost, &e.PricePerL, &e.Kilometers,
		&e.FuelledAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func scanRows(rows pgx.Rows) ([]*fuelentry.FuelEntry, error) {
	entries := make([]*fuelentry.FuelEntry, 0)
	for rows.Next() {
		var e fuelentry.FuelEntry
		if err := rows.Scan(&e.ID, &e.Liters, &e.TotalCost, &e.PricePerL, &e.Kilometers,
			&e.FuelledAt, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
