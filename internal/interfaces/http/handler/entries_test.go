// ABOUTME: E2E HTTP tests for all /v1/entries endpoints.
// ABOUTME: Uses fiber.App.Test() — no real server or port needed.

package handler_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	appfuelentry "github.com/amavis442/gasolina-api/internal/application/fuelentry"
	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

func validCreateBody() map[string]any {
	return map[string]any{
		"id":         "entry-1",
		"liters":     40.5,
		"total_cost": 65.80,
		"price_per_l": 1.625,
		"kilometers":   123456.7,
		"fuelled_at": time.Now().Format(time.RFC3339),
	}
}

// GET /v1/entries

func TestGetAll_EmptyRepo_ReturnsEmptyArray(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodGet, "/v1/entries", nil, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAll_WithEntries_ReturnsThem(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	seedEntry(repo, "entry-2")
	app := newApp(repo)

	resp := do(app, http.MethodGet, "/v1/entries", nil, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entries []fuelentry.FuelEntry
	decode(resp, &entries)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetAll_WithSinceParam_FiltersEntries(t *testing.T) {
	repo := newMockRepo()
	old := seedEntry(repo, "old-entry")
	old.UpdatedAt = time.Now().Add(-48 * time.Hour)
	recent := seedEntry(repo, "recent-entry")
	recent.UpdatedAt = time.Now()
	app := newApp(repo)

	cutoffMs := time.Now().Add(-1 * time.Hour).UnixMilli()
	resp := do(app, http.MethodGet, fmt.Sprintf("/v1/entries?since=%d", cutoffMs), nil, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entries []fuelentry.FuelEntry
	decode(resp, &entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after cutoff, got %d", len(entries))
	}
}

func TestGetAll_InvalidSinceParam_Returns400(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodGet, "/v1/entries?since=not-a-number", nil, validToken())

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// GET /v1/entries/:id

func TestGetByID_ExistingEntry_Returns200(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	app := newApp(repo)

	resp := do(app, http.MethodGet, "/v1/entries/entry-1", nil, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entry fuelentry.FuelEntry
	decode(resp, &entry)
	if entry.ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %q", entry.ID)
	}
}

func TestGetByID_MissingEntry_Returns404(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodGet, "/v1/entries/does-not-exist", nil, validToken())

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// POST /v1/entries

func TestCreate_ValidBody_Returns201(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodPost, "/v1/entries", validCreateBody(), validToken())

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var entry fuelentry.FuelEntry
	decode(resp, &entry)
	if entry.ID != "entry-1" {
		t.Errorf("expected ID entry-1, got %q", entry.ID)
	}
}

func TestCreate_InvalidBody_Returns422(t *testing.T) {
	app := newApp(newMockRepo())

	// Missing required id field
	resp := do(app, http.MethodPost, "/v1/entries", map[string]any{
		"liters":     40.5,
		"total_cost": 65.80,
		"fuelled_at": time.Now().Format(time.RFC3339),
	}, validToken())

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}

// PUT /v1/entries/:id

func TestUpdate_ExistingEntry_Returns200(t *testing.T) {
	repo := newMockRepo()
	existing := seedEntry(repo, "entry-1")
	existing.UpdatedAt = time.Now().Add(-1 * time.Hour)
	app := newApp(repo)

	body := map[string]any{
		"liters":     99.9,
		"total_cost": 150.0,
		"price_per_l": 1.5,
		"kilometers":   6000,
		"fuelled_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	resp := do(app, http.MethodPut, "/v1/entries/entry-1", body, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdate_MissingEntry_Returns422(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodPut, "/v1/entries/does-not-exist", map[string]any{
		"liters":     10.0,
		"total_cost": 20.0,
		"fuelled_at": time.Now().Format(time.RFC3339),
	}, validToken())

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}

// DELETE /v1/entries/:id

func TestDelete_ExistingEntry_Returns204(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	app := newApp(repo)

	resp := do(app, http.MethodDelete, "/v1/entries/entry-1", nil, validToken())

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDelete_MissingEntry_Returns500(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodDelete, "/v1/entries/does-not-exist", nil, validToken())

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}

// POST /v1/entries/sync

func TestSync_ValidPayload_Returns200(t *testing.T) {
	app := newApp(newMockRepo())

	body := appfuelentry.SyncInput{
		LastSyncAt: 0,
		Entries: []appfuelentry.SyncEntry{
			{
				ID:        "entry-1",
				Liters:    40.0,
				TotalCost: 60.0,
				PricePerL: 1.5,
				Kilometers: 1000,
				FuelledAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	resp := do(app, http.MethodPost, "/v1/entries/sync", body, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entries []fuelentry.FuelEntry
	decode(resp, &entries)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry after sync, got %d", len(entries))
	}
}

func TestSync_EmptyPayload_Returns200WithAllEntries(t *testing.T) {
	repo := newMockRepo()
	seedEntry(repo, "entry-1")
	seedEntry(repo, "entry-2")
	app := newApp(repo)

	resp := do(app, http.MethodPost, "/v1/entries/sync", appfuelentry.SyncInput{
		LastSyncAt: 0,
	}, validToken())

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entries []fuelentry.FuelEntry
	decode(resp, &entries)
	if len(entries) != 2 {
		t.Errorf("expected full recovery to return 2 entries, got %d", len(entries))
	}
}
