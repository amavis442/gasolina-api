// ABOUTME: E2E HTTP tests for the POST /auth/token endpoint.
// ABOUTME: Uses fiber.App.Test() — no real server or port needed.

package handler_test

import (
	"net/http"
	"testing"
)

func TestAuthToken_ValidSecret_ReturnsToken(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodPost, "/auth/token", map[string]string{
		"device_secret": testDeviceSecret,
	}, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	decode(resp, &body)
	if body["token"] == "" {
		t.Error("expected token in response body, got empty string")
	}
}

func TestAuthToken_WrongSecret_Returns401(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodPost, "/auth/token", map[string]string{
		"device_secret": "wrong-secret",
	}, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthToken_InvalidBody_Returns400(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodPost, "/auth/token", nil, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestProtectedRoute_NoToken_Returns401(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodGet, "/v1/entries", nil, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", resp.StatusCode)
	}
}

func TestProtectedRoute_InvalidToken_Returns401(t *testing.T) {
	app := newApp(newMockRepo())

	resp := do(app, http.MethodGet, "/v1/entries", nil, "not-a-valid-jwt")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", resp.StatusCode)
	}
}

func TestAuthToken_RefreshTokenHeader_PresentOnAuthenticatedResponse(t *testing.T) {
	repo := newMockRepo()
	app := newApp(repo)

	resp := do(app, http.MethodGet, "/v1/entries", nil, validToken())

	if resp.Header.Get("X-Refresh-Token") == "" {
		t.Error("expected X-Refresh-Token header on authenticated response")
	}
}
