// ABOUTME: Fiber app wiring — registers all routes and middleware.
// ABOUTME: Called from cmd/api/main.go; returns a configured *fiber.App.

package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberswagger "github.com/gofiber/swagger"
	"github.com/amavis442/gasolina-api/internal/interfaces/http/handler"
	"github.com/amavis442/gasolina-api/internal/interfaces/http/middleware"
)

func New(
	authHandler *handler.AuthHandler,
	entriesHandler *handler.EntriesHandler,
	jwtSecret string,
	tokenTTL time.Duration,
) *fiber.App {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/swagger/*", fiberswagger.HandlerDefault)

	app.Post("/auth/token", authHandler.Token)

	v1 := app.Group("/v1", middleware.JWT(jwtSecret, tokenTTL))
	v1.Get("/entries", entriesHandler.GetAll)
	v1.Post("/entries", entriesHandler.Create)
	v1.Post("/entries/sync", entriesHandler.Sync)
	v1.Get("/entries/:id", entriesHandler.GetByID)
	v1.Put("/entries/:id", entriesHandler.Update)
	v1.Delete("/entries/:id", entriesHandler.Delete)

	return app
}
