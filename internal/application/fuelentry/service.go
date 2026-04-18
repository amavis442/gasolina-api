// ABOUTME: Application service orchestrating FuelEntry use cases.
// ABOUTME: Delegates persistence to the domain repository interface.

package fuelentry

import (
	"context"
	"time"

	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

type Service struct {
	repo fuelentry.Repository
}

func NewService(repo fuelentry.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Add(ctx context.Context, in CreateInput) (*fuelentry.FuelEntry, error) {
	now := time.Now().UTC()
	entry := &fuelentry.FuelEntry{
		ID:        in.ID,
		Liters:    in.Liters,
		TotalCost: in.TotalCost,
		PricePerL: in.PricePerL,
		Odometer:  in.Odometer,
		FuelledAt: in.FuelledAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := entry.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *Service) Update(ctx context.Context, in UpdateInput) (*fuelentry.FuelEntry, error) {
	existing, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	// Last-write-wins on updated_at
	if !in.UpdatedAt.IsZero() && in.UpdatedAt.Before(existing.UpdatedAt) {
		return existing, nil
	}
	existing.Liters = in.Liters
	existing.TotalCost = in.TotalCost
	existing.PricePerL = in.PricePerL
	existing.Odometer = in.Odometer
	existing.FuelledAt = in.FuelledAt
	existing.UpdatedAt = time.Now().UTC()
	if err := existing.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id, time.Now().UTC())
}

func (s *Service) GetByID(ctx context.Context, id string) (*fuelentry.FuelEntry, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, since *time.Time) ([]*fuelentry.FuelEntry, error) {
	if since != nil {
		return s.repo.GetSince(ctx, *since)
	}
	return s.repo.GetAll(ctx)
}

func (s *Service) Sync(ctx context.Context, in SyncInput) ([]*fuelentry.FuelEntry, error) {
	var sinceTime *time.Time
	if in.LastSyncAt > 0 {
		t := time.UnixMilli(in.LastSyncAt).UTC()
		sinceTime = &t
	}

	for _, se := range in.Entries {
		entry := &fuelentry.FuelEntry{
			ID:        se.ID,
			Liters:    se.Liters,
			TotalCost: se.TotalCost,
			PricePerL: se.PricePerL,
			Odometer:  se.Odometer,
			FuelledAt: se.FuelledAt,
			UpdatedAt: se.UpdatedAt,
			DeletedAt: se.DeletedAt,
		}
		existing, err := s.repo.GetByID(ctx, se.ID)
		if err != nil {
			// Not found — insert
			entry.CreatedAt = time.Now().UTC()
			if saveErr := s.repo.Save(ctx, entry); saveErr != nil {
				return nil, saveErr
			}
			continue
		}
		if se.UpdatedAt.After(existing.UpdatedAt) {
			entry.CreatedAt = existing.CreatedAt
			if se.DeletedAt != nil {
				if delErr := s.repo.Delete(ctx, se.ID, *se.DeletedAt); delErr != nil {
					return nil, delErr
				}
			} else {
				if upErr := s.repo.Update(ctx, entry); upErr != nil {
					return nil, upErr
				}
			}
		}
	}

	return s.GetAll(ctx, sinceTime)
}
