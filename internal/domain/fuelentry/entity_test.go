// ABOUTME: Unit tests for the FuelEntry domain entity and its Validate() method.
// ABOUTME: No I/O — pure domain logic only.

package fuelentry_test

import (
	"testing"
	"time"

	"github.com/amavis442/gasolina-api/internal/domain/fuelentry"
)

func validEntry() *fuelentry.FuelEntry {
	return &fuelentry.FuelEntry{
		ID:        "entry-1",
		Liters:    40.5,
		TotalCost: 65.80,
		PricePerL: 1.625,
		Kilometers: 123456.7,
		FuelledAt: time.Now(),
	}
}

func TestValidate_ValidEntry(t *testing.T) {
	if err := validEntry().Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_MissingID(t *testing.T) {
	e := validEntry()
	e.ID = ""
	assertValidateError(t, e, "id is required")
}

func TestValidate_ZeroLiters(t *testing.T) {
	e := validEntry()
	e.Liters = 0
	assertValidateError(t, e, "liters must be positive")
}

func TestValidate_NegativeLiters(t *testing.T) {
	e := validEntry()
	e.Liters = -1
	assertValidateError(t, e, "liters must be positive")
}

func TestValidate_ZeroTotalCost(t *testing.T) {
	e := validEntry()
	e.TotalCost = 0
	assertValidateError(t, e, "total_cost must be positive")
}

func TestValidate_NegativeTotalCost(t *testing.T) {
	e := validEntry()
	e.TotalCost = -0.01
	assertValidateError(t, e, "total_cost must be positive")
}

func TestValidate_NegativeKilometers(t *testing.T) {
	e := validEntry()
	e.Kilometers = -1
	assertValidateError(t, e, "kilometers must be non-negative")
}

func TestValidate_ZeroKilometersIsAllowed(t *testing.T) {
	e := validEntry()
	e.Kilometers = 0
	if err := e.Validate(); err != nil {
		t.Errorf("expected zero kilometers to be valid, got: %v", err)
	}
}

func TestValidate_ZeroFuelledAt(t *testing.T) {
	e := validEntry()
	e.FuelledAt = time.Time{}
	assertValidateError(t, e, "fuelled_at is required")
}

func TestValidate_DeletedAtIsOptional(t *testing.T) {
	e := validEntry()
	e.DeletedAt = nil
	if err := e.Validate(); err != nil {
		t.Errorf("expected nil deleted_at to be valid, got: %v", err)
	}
}

func assertValidateError(t *testing.T, e *fuelentry.FuelEntry, want string) {
	t.Helper()
	err := e.Validate()
	if err == nil {
		t.Fatalf("expected error %q, got nil", want)
	}
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}
