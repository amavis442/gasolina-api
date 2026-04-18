// ABOUTME: HTTP handlers for FuelEntry CRUD and sync endpoints using Fiber.
// ABOUTME: Delegates all business logic to the application service.

package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	appfuelentry "github.com/amavis442/gasolina-api/internal/application/fuelentry"
)

type EntriesHandler struct {
	svc *appfuelentry.Service
}

func NewEntriesHandler(svc *appfuelentry.Service) *EntriesHandler {
	return &EntriesHandler{svc: svc}
}

// GetAll godoc
// @Summary      List fuel entries
// @Description  Returns all fuel entries. Add ?since=<unix_ms> for a delta pull of entries updated after that timestamp.
// @Tags         entries
// @Produce      json
// @Param        since  query     int  false  "Unix timestamp in milliseconds; only entries updated after this value are returned"
// @Success      200    {array}   fuelEntryResponse
// @Failure      400    {object}  errorResponse
// @Failure      500    {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries [get]
func (h *EntriesHandler) GetAll(c *fiber.Ctx) error {
	var since *time.Time
	if s := c.Query("since"); s != "" {
		ms, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid since parameter"})
		}
		t := time.UnixMilli(ms).UTC()
		since = &t
	}
	entries, err := h.svc.GetAll(c.Context(), since)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(entries)
}

// GetByID godoc
// @Summary      Get a fuel entry
// @Description  Returns a single fuel entry by its ID.
// @Tags         entries
// @Produce      json
// @Param        id   path      string  true  "Entry ID"
// @Success      200  {object}  fuelEntryResponse
// @Failure      404  {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries/{id} [get]
func (h *EntriesHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	entry, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "entry not found"})
	}
	return c.Status(fiber.StatusOK).JSON(entry)
}

// Create godoc
// @Summary      Create a fuel entry
// @Description  Pushes a new fuel entry to the server.
// @Tags         entries
// @Accept       json
// @Produce      json
// @Param        body  body      createEntryRequest  true  "Fuel entry payload"
// @Success      201   {object}  fuelEntryResponse
// @Failure      400   {object}  errorResponse
// @Failure      422   {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries [post]
func (h *EntriesHandler) Create(c *fiber.Ctx) error {
	var in appfuelentry.CreateInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	entry, err := h.svc.Add(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(entry)
}

// Update godoc
// @Summary      Update a fuel entry
// @Description  Pushes an update for an existing entry. Last-write-wins on updated_at — if the incoming updated_at is older than the stored value the update is silently ignored.
// @Tags         entries
// @Accept       json
// @Produce      json
// @Param        id    path      string              true  "Entry ID"
// @Param        body  body      updateEntryRequest  true  "Updated fuel entry payload"
// @Success      200   {object}  fuelEntryResponse
// @Failure      400   {object}  errorResponse
// @Failure      422   {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries/{id} [put]
func (h *EntriesHandler) Update(c *fiber.Ctx) error {
	var in appfuelentry.UpdateInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	in.ID = c.Params("id")
	entry, err := h.svc.Update(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(entry)
}

// Delete godoc
// @Summary      Delete a fuel entry
// @Description  Soft-deletes an entry by setting its deleted_at timestamp. Deleted entries are included in sync responses so clients can propagate the deletion.
// @Tags         entries
// @Produce      json
// @Param        id   path  string  true  "Entry ID"
// @Success      204
// @Failure      500  {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries/{id} [delete]
func (h *EntriesHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Sync godoc
// @Summary      Bidirectional bulk sync
// @Description  Merges the client's entries with the server's. Pass last_sync_at: 0 for a full recovery pull. The server applies last-write-wins on each entry and returns all entries updated since last_sync_at.
// @Tags         entries
// @Accept       json
// @Produce      json
// @Param        body  body      syncRequest  true  "Sync payload"
// @Success      200   {array}   fuelEntryResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Security     BearerAuth
// @Router       /v1/entries/sync [post]
func (h *EntriesHandler) Sync(c *fiber.Ctx) error {
	var in appfuelentry.SyncInput
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	entries, err := h.svc.Sync(c.Context(), in)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(entries)
}
