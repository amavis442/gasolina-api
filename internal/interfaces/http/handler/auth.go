// ABOUTME: HTTP handler for device authentication — exchanges device secret for a JWT.
// ABOUTME: POST /auth/token endpoint.

package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/amavis442/gasolina-api/auth"
)

type AuthHandler struct {
	deviceSecret string
	jwtSecret    string
	tokenTTL     time.Duration
}

func NewAuthHandler(deviceSecret, jwtSecret string, tokenTTL time.Duration) *AuthHandler {
	return &AuthHandler{deviceSecret: deviceSecret, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

// Token godoc
// @Summary      Get JWT token
// @Description  Exchange the device secret for a signed JWT. Store the returned token and send it as "Bearer <token>" on all /v1/* requests.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      tokenRequest   true  "Device secret"
// @Success      200   {object}  tokenResponse
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /auth/token [post]
func (h *AuthHandler) Token(c *fiber.Ctx) error {
	var body struct {
		DeviceSecret string `json:"device_secret"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if body.DeviceSecret != h.deviceSecret {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid device secret"})
	}
	token, err := auth.GenerateToken(h.jwtSecret, h.tokenTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
}

// tokenRequest is the body for POST /auth/token.
type tokenRequest struct {
	DeviceSecret string `json:"device_secret" example:"my-device-secret"`
}

// tokenResponse is returned by POST /auth/token.
type tokenResponse struct {
	Token string `json:"token" example:"eyJhbGci..."`
}

// errorResponse is the standard error envelope.
type errorResponse struct {
	Error string `json:"error" example:"invalid device secret"`
}
