// ABOUTME: Swagger-only request/response type definitions for handler annotations.
// ABOUTME: Avoids package alias conflicts between domain and application fuelentry packages.

package handler

import "time"

// createEntryRequest is the body for POST /v1/entries.
type createEntryRequest struct {
	ID        string    `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Liters    float64   `json:"liters"      example:"40.5"`
	TotalCost float64   `json:"total_cost"  example:"65.80"`
	PricePerL float64   `json:"price_per_l" example:"1.625"`
	Odometer  float64   `json:"odometer"    example:"123456.7"`
	FuelledAt time.Time `json:"fuelled_at"  example:"2024-06-01T08:00:00Z"`
}

// updateEntryRequest is the body for PUT /v1/entries/{id}.
type updateEntryRequest struct {
	Liters    float64   `json:"liters"      example:"40.5"`
	TotalCost float64   `json:"total_cost"  example:"65.80"`
	PricePerL float64   `json:"price_per_l" example:"1.625"`
	Odometer  float64   `json:"odometer"    example:"123456.7"`
	FuelledAt time.Time `json:"fuelled_at"  example:"2024-06-01T08:00:00Z"`
	UpdatedAt time.Time `json:"updated_at"  example:"2024-06-01T09:00:00Z"`
}

// syncEntryItem is one entry inside a sync payload.
type syncEntryItem struct {
	ID        string     `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Liters    float64    `json:"liters"      example:"40.5"`
	TotalCost float64    `json:"total_cost"  example:"65.80"`
	PricePerL float64    `json:"price_per_l" example:"1.625"`
	Odometer  float64    `json:"odometer"    example:"123456.7"`
	FuelledAt time.Time  `json:"fuelled_at"  example:"2024-06-01T08:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at"  example:"2024-06-01T09:00:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// syncRequest is the body for POST /v1/entries/sync.
type syncRequest struct {
	LastSyncAt int64           `json:"last_sync_at" example:"0"`
	Entries    []syncEntryItem `json:"entries"`
}

// fuelEntryResponse mirrors domain.FuelEntry for swagger output docs.
type fuelEntryResponse struct {
	ID        string     `json:"id"          example:"550e8400-e29b-41d4-a716-446655440000"`
	Liters    float64    `json:"liters"      example:"40.5"`
	TotalCost float64    `json:"total_cost"  example:"65.80"`
	PricePerL float64    `json:"price_per_l" example:"1.625"`
	Odometer  float64    `json:"odometer"    example:"123456.7"`
	FuelledAt time.Time  `json:"fuelled_at"  example:"2024-06-01T08:00:00Z"`
	CreatedAt time.Time  `json:"created_at"  example:"2024-06-01T07:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at"  example:"2024-06-01T09:00:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
