// ABOUTME: FuelEntry domain entity and validation logic.
// ABOUTME: Zero external imports — pure domain layer.

package fuelentry

import (
	"errors"
	"time"
)

type FuelEntry struct {
	ID          string    `json:"id"`
	Liters      float64   `json:"liters"`
	TotalCost   float64   `json:"total_cost"`
	PricePerL   float64   `json:"price_per_l"`
	Odometer    float64   `json:"odometer"`
	FuelledAt   time.Time `json:"fuelled_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (e *FuelEntry) Validate() error {
	if e.ID == "" {
		return errors.New("id is required")
	}
	if e.Liters <= 0 {
		return errors.New("liters must be positive")
	}
	if e.TotalCost <= 0 {
		return errors.New("total_cost must be positive")
	}
	if e.Odometer < 0 {
		return errors.New("odometer must be non-negative")
	}
	if e.FuelledAt.IsZero() {
		return errors.New("fuelled_at is required")
	}
	return nil
}
