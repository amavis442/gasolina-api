// ABOUTME: Unit tests for the FuelEntry application service.
// ABOUTME: Uses an in-memory mock repository — no I/O.

package fuelentry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appfuelentry "github.com/amavis442/gasolina-api/internal/application/fuelentry"
	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

var ctx = context.Background()

// helpers

func validCreateInput() appfuelentry.CreateInput {
	return appfuelentry.CreateInput{
		ID:        "entry-1",
		Liters:    40.5,
		TotalCost: 65.80,
		PricePerL: 1.625,
		Kilometers:  123456.7,
		FuelledAt: time.Now(),
	}
}

func seedEntry(repo *mockRepo, id string) *fuelentry.FuelEntry {
	e := &fuelentry.FuelEntry{
		ID:        id,
		Liters:    30.0,
		TotalCost: 50.0,
		PricePerL: 1.66,
		Kilometers:  5000,
		FuelledAt: time.Now().Add(-24 * time.Hour),
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	repo.entries[id] = e
	return e
}

// Add

func TestAdd_CreatesEntry(t *testing.T) {
	svc := appfuelentry.NewService(newMockRepo())
	in := validCreateInput()

	entry, err := svc.Add(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != in.ID {
		t.Errorf("expected ID %q, got %q", in.ID, entry.ID)
	}
	if entry.Liters != in.Liters {
		t.Errorf("expected Liters %v, got %v", in.Liters, entry.Liters)
	}
	if entry.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if entry.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestAdd_ReturnsValidationError(t *testing.T) {
	svc := appfuelentry.NewService(newMockRepo())
	in := validCreateInput()
	in.ID = ""

	_, err := svc.Add(ctx, in)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestAdd_ReturnsRepoError(t *testing.T) {
	repoErr := errors.New("db unavailable")
	repo := newMockRepo()
	repo.saveFn = func(_ *fuelentry.FuelEntry) error { return repoErr }
	svc := appfuelentry.NewService(repo)

	_, err := svc.Add(ctx, validCreateInput())

	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}

// GetByID

func TestGetByID_ReturnsEntry(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	svc := appfuelentry.NewService(repo)

	entry, err := svc.GetByID(ctx, "entry-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %q", entry.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := appfuelentry.NewService(newMockRepo())

	_, err := svc.GetByID(ctx, "missing")

	if err == nil {
		t.Fatal("expected error for missing entry, got nil")
	}
}

// GetAll

func TestGetAll_NoFilter_ReturnsAll(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	seedEntry(repo, "entry-2")
	svc := appfuelentry.NewService(repo)

	entries, err := svc.GetAll(ctx, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetAll_WithSince_FiltersOldEntries(t *testing.T) {
	repo := newMockRepo()
	old := seedEntry(repo, "old-entry")
	old.UpdatedAt = time.Now().Add(-48 * time.Hour)

	recent := seedEntry(repo, "recent-entry")
	recent.UpdatedAt = time.Now()

	cutoff := time.Now().Add(-1 * time.Hour)
	svc := appfuelentry.NewService(repo)

	entries, err := svc.GetAll(ctx, &cutoff)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after cutoff, got %d", len(entries))
	}
	if entries[0].ID != "recent-entry" {
		t.Errorf("expected recent-entry, got %q", entries[0].ID)
	}
}

// Update

func TestUpdate_LastWriteWins_NewerWins(t *testing.T) {
	repo := newMockRepo()
	existing := seedEntry(repo, "entry-1")
	existing.UpdatedAt = time.Now().Add(-1 * time.Hour)
	svc := appfuelentry.NewService(repo)

	in := appfuelentry.UpdateInput{
		ID:        "entry-1",
		Liters:    99.9,
		TotalCost: 150.0,
		PricePerL: 1.5,
		Kilometers:  6000,
		FuelledAt: time.Now(),
		UpdatedAt: time.Now(), // newer than existing
	}

	entry, err := svc.Update(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Liters != 99.9 {
		t.Errorf("expected Liters 99.9, got %v", entry.Liters)
	}
}

func TestUpdate_LastWriteWins_OlderLoses(t *testing.T) {
	repo := newMockRepo()
	existing := seedEntry(repo, "entry-1")
	existing.UpdatedAt = time.Now()
	svc := appfuelentry.NewService(repo)

	in := appfuelentry.UpdateInput{
		ID:        "entry-1",
		Liters:    99.9,
		TotalCost: 150.0,
		PricePerL: 1.5,
		Kilometers:  6000,
		FuelledAt: time.Now(),
		UpdatedAt: time.Now().Add(-2 * time.Hour), // older than existing
	}

	entry, err := svc.Update(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return existing unchanged
	if entry.Liters != existing.Liters {
		t.Errorf("expected Liters %v (unchanged), got %v", existing.Liters, entry.Liters)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := appfuelentry.NewService(newMockRepo())

	_, err := svc.Update(ctx, appfuelentry.UpdateInput{ID: "missing"})

	if err == nil {
		t.Fatal("expected error for missing entry, got nil")
	}
}

// Delete

func TestDelete_SoftDeletesEntry(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	svc := appfuelentry.NewService(repo)

	err := svc.Delete(ctx, "entry-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := repo.entries["entry-1"]
	if e.DeletedAt == nil {
		t.Error("expected DeletedAt to be set after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := appfuelentry.NewService(newMockRepo())

	err := svc.Delete(ctx, "missing")

	if err == nil {
		t.Fatal("expected error for missing entry, got nil")
	}
}

// Sync

func TestSync_InsertsNewEntries(t *testing.T) {
	repo := newMockRepo()
	svc := appfuelentry.NewService(repo)

	in := appfuelentry.SyncInput{
		LastSyncAt: 0,
		Entries: []appfuelentry.SyncEntry{
			{
				ID:        "entry-1",
				Liters:    40.0,
				TotalCost: 60.0,
				PricePerL: 1.5,
				Kilometers:  1000,
				FuelledAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	entries, err := svc.Sync(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after sync, got %d", len(entries))
	}
}

func TestSync_UpdatesExistingWhenNewer(t *testing.T) {
	repo := newMockRepo()
	existing := seedEntry(repo, "entry-1")
	existing.UpdatedAt = time.Now().Add(-1 * time.Hour)
	svc := appfuelentry.NewService(repo)

	newerTime := time.Now()
	in := appfuelentry.SyncInput{
		Entries: []appfuelentry.SyncEntry{
			{
				ID:        "entry-1",
				Liters:    99.0,
				TotalCost: 150.0,
				PricePerL: 1.51,
				Kilometers:  9999,
				FuelledAt: time.Now(),
				UpdatedAt: newerTime,
			},
		},
	}

	_, err := svc.Sync(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.entries["entry-1"].Liters != 99.0 {
		t.Errorf("expected updated Liters 99.0, got %v", repo.entries["entry-1"].Liters)
	}
}

func TestSync_SkipsExistingWhenOlder(t *testing.T) {
	repo := newMockRepo()
	existing := seedEntry(repo, "entry-1")
	existing.UpdatedAt = time.Now()
	svc := appfuelentry.NewService(repo)

	in := appfuelentry.SyncInput{
		Entries: []appfuelentry.SyncEntry{
			{
				ID:        "entry-1",
				Liters:    99.0,
				TotalCost: 150.0,
				PricePerL: 1.51,
				Kilometers:  9999,
				FuelledAt: time.Now(),
				UpdatedAt: time.Now().Add(-2 * time.Hour), // older
			},
		},
	}

	_, err := svc.Sync(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.entries["entry-1"].Liters != existing.Liters {
		t.Errorf("expected Liters unchanged at %v, got %v", existing.Liters, repo.entries["entry-1"].Liters)
	}
}

func TestSync_SoftDeletesWhenNewerAndDeletedAtSet(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	svc := appfuelentry.NewService(repo)

	deletedAt := time.Now()
	in := appfuelentry.SyncInput{
		Entries: []appfuelentry.SyncEntry{
			{
				ID:        "entry-1",
				Liters:    30.0,
				TotalCost: 50.0,
				PricePerL: 1.66,
				Kilometers:  5000,
				FuelledAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: &deletedAt,
			},
		},
	}

	_, err := svc.Sync(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.entries["entry-1"].DeletedAt == nil {
		t.Error("expected DeletedAt to be set after sync delete")
	}
}

func TestSync_FullRecovery_LastSyncAtZero(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	seedEntry(repo, "entry-2")
	svc := appfuelentry.NewService(repo)

	in := appfuelentry.SyncInput{LastSyncAt: 0, Entries: nil}

	entries, err := svc.Sync(ctx, in)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected full recovery to return 2 entries, got %d", len(entries))
	}
}
