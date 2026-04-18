// ABOUTME: In-memory mock repository implementing domain.Repository for service unit tests.
// ABOUTME: No I/O — all state is held in a map.

package fuelentry_test

import (
	"context"
	"errors"
	"time"

	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

var errNotFound = errors.New("not found")

type mockRepo struct {
	entries map[string]*fuelentry.FuelEntry
	saveFn  func(*fuelentry.FuelEntry) error
}

func newMockRepo() *mockRepo {
	return &mockRepo{entries: make(map[string]*fuelentry.FuelEntry)}
}

func (m *mockRepo) GetAll(_ context.Context) ([]*fuelentry.FuelEntry, error) {
	out := make([]*fuelentry.FuelEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out, nil
}

func (m *mockRepo) GetSince(_ context.Context, since time.Time) ([]*fuelentry.FuelEntry, error) {
	var out []*fuelentry.FuelEntry
	for _, e := range m.entries {
		if e.UpdatedAt.After(since) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *mockRepo) GetByID(_ context.Context, id string) (*fuelentry.FuelEntry, error) {
	e, ok := m.entries[id]
	if !ok {
		return nil, errNotFound
	}
	return e, nil
}

func (m *mockRepo) Save(_ context.Context, e *fuelentry.FuelEntry) error {
	if m.saveFn != nil {
		return m.saveFn(e)
	}
	m.entries[e.ID] = e
	return nil
}

func (m *mockRepo) Update(_ context.Context, e *fuelentry.FuelEntry) error {
	if _, ok := m.entries[e.ID]; !ok {
		return errNotFound
	}
	m.entries[e.ID] = e
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id string, deletedAt time.Time) error {
	e, ok := m.entries[id]
	if !ok {
		return errNotFound
	}
	e.DeletedAt = &deletedAt
	e.UpdatedAt = deletedAt
	return nil
}
