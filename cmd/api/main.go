// ABOUTME: Application entry point — wires all layers and starts the HTTP server.
// ABOUTME: The only place in the codebase allowed to import across all layers.

// @title           Gasolina API
// @version         1.0
// @description     Cloud sync backend for the Gasolina fuel tracking app.
// @contact.name    Patrick
// @contact.url     https://github.com/amavis442/gasolina-api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter: Bearer <token>

// @host      localhost:8080
// @BasePath  /

package main

import (
	_ "github.com/amavis442/gasolina-api/docs"
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/amavis442/gasolina-api/config"
	appfuelentry "github.com/amavis442/gasolina-api/internal/application/fuelentry"
	"github.com/amavis442/gasolina-api/internal/infrastructure/persistence/postgres"
	"github.com/amavis442/gasolina-api/internal/interfaces/http/handler"
	"github.com/amavis442/gasolina-api/server"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ttl, err := time.ParseDuration(cfg.TokenTTL)
	if err != nil {
		log.Fatalf("invalid token_ttl: %v", err)
	}

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := postgres.NewFuelEntryRepository(db)
	svc := appfuelentry.NewService(repo)

	authHandler := handler.NewAuthHandler(cfg.DeviceSecret, cfg.JWTSecret, ttl)
	entriesHandler := handler.NewEntriesHandler(svc)

	app := server.New(authHandler, entriesHandler, cfg.JWTSecret, ttl)

	log.Printf("gasolina-api listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
