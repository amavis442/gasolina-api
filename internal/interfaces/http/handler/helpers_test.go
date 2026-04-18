// ABOUTME: Test helpers — in-memory mock repo, app factory, and request builders for handler e2e tests.
// ABOUTME: Uses fiber.App.Test() so no real server or port is needed.

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/amavis442/gasolina-api/auth"
	appfuelentry "github.com/amavis442/gasolina-api/internal/application/fuelentry"
	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
	"github.com/amavis442/gasolina-api/internal/interfaces/http/handler"
	"github.com/amavis442/gasolina-api/internal/interfaces/http/middleware"
)

const (
	testDeviceSecret = "test-device-secret"
	testJWTSecret    = "test-jwt-secret"
	testTokenTTL     = time.Hour
)

// errNotFound is returned by the mock when an entry is absent.
var errNotFound = errors.New("not found")

// mockRepo is an in-memory implementation of domain.Repository.
type mockRepo struct {
	entries map[string]*fuelentry.FuelEntry
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

// newApp wires up the full Fiber app with the given repo.
func newApp(repo *mockRepo) *fiber.App {
	svc := appfuelentry.NewService(repo)
	authH := handler.NewAuthHandler(testDeviceSecret, testJWTSecret, testTokenTTL)
	entriesH := handler.NewEntriesHandler(svc)

	app := fiber.New()
	app.Post("/auth/token", authH.Token)
	v1 := app.Group("/v1", middleware.JWT(testJWTSecret, testTokenTTL))
	v1.Get("/entries", entriesH.GetAll)
	v1.Post("/entries", entriesH.Create)
	v1.Post("/entries/sync", entriesH.Sync)
	v1.Get("/entries/:id", entriesH.GetByID)
	v1.Put("/entries/:id", entriesH.Update)
	v1.Delete("/entries/:id", entriesH.Delete)

	return app
}

// validToken returns a signed JWT for use in authenticated requests.
func validToken() string {
	t, _ := auth.GenerateToken(testJWTSecret, testTokenTTL)
	return t
}

// do executes a request against the Fiber app via app.Test().
func do(app *fiber.App, method, path string, body any, token string) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req)
	return resp
}

// decode reads JSON from an http.Response body into v.
func decode(resp *http.Response, v any) {
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(v)
}

// seedEntry adds a pre-built entry directly into the mock repo.
func seedEntry(repo *mockRepo, id string) *fuelentry.FuelEntry {
	e := &fuelentry.FuelEntry{
		ID:        id,
		Liters:    30.0,
		TotalCost: 50.0,
		PricePerL: 1.66,
		Kilometers: 5000,
		FuelledAt: time.Now().Add(-24 * time.Hour),
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	repo.entries[id] = e
	return e
}
