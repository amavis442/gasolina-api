// ABOUTME: Input and output DTOs for FuelEntry use cases.
// ABOUTME: Decouples application layer from transport/domain representations.

package fuelentry

import "time"

type CreateInput struct {
	ID        string    `json:"id"`
	Liters    float64   `json:"liters"`
	TotalCost float64   `json:"total_cost"`
	PricePerL float64   `json:"price_per_l"`
	Odometer  float64   `json:"odometer"`
	FuelledAt time.Time `json:"fuelled_at"`
}

type UpdateInput struct {
	ID        string    `json:"id"`
	Liters    float64   `json:"liters"`
	TotalCost float64   `json:"total_cost"`
	PricePerL float64   `json:"price_per_l"`
	Odometer  float64   `json:"odometer"`
	FuelledAt time.Time `json:"fuelled_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SyncInput struct {
	LastSyncAt int64        `json:"last_sync_at"` // Unix ms; 0 = full recovery
	Entries    []SyncEntry  `json:"entries"`
}

type SyncEntry struct {
	ID        string     `json:"id"`
	Liters    float64    `json:"liters"`
	TotalCost float64    `json:"total_cost"`
	PricePerL float64    `json:"price_per_l"`
	Odometer  float64    `json:"odometer"`
	FuelledAt time.Time  `json:"fuelled_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
