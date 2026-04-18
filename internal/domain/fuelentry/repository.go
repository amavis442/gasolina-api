// ABOUTME: Repository interface (port) for FuelEntry persistence.
// ABOUTME: Defines the contract; implementations live in infrastructure/.

package fuelentry

import (
	"context"
	"time"
)

type Repository interface {
	GetAll(ctx context.Context) ([]*FuelEntry, error)
	GetSince(ctx context.Context, since time.Time) ([]*FuelEntry, error)
	GetByID(ctx context.Context, id string) (*FuelEntry, error)
	Save(ctx context.Context, entry *FuelEntry) error
	Update(ctx context.Context, entry *FuelEntry) error
	Delete(ctx context.Context, id string, deletedAt time.Time) error
}
