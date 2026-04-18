// ABOUTME: E2E tests for /v1/entries — CRUD operations including edge cases.
// ABOUTME: Tests run sequentially and share state via module-level variables.

import { assertEquals, assertExists, assert } from "@std/assert";
import { ApiClient } from "../client.ts";
import { cleanTestData } from "../cleanup.ts";

interface FuelEntry {
  id: string;
  liters: number;
  total_cost: number;
  price_per_l: number;
  odometer: number;
  fuelled_at: string;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

interface ErrorBody {
  error: string;
}

// Shared state across sequential tests
const state = {
  entryId: "",
  createdAt: "",
  updatedAt: "",
};

const ENTRY_ID = `e2e-test-${Date.now()}`;

const client = new ApiClient();

Deno.test("setup — remove leftover test rows from previous runs", async () => {
  await cleanTestData();
});

// ---------- CREATE ----------

Deno.test("POST /v1/entries — creates a new entry and returns 201", async () => {
  const res = await client.post<FuelEntry>("/v1/entries", {
    id: ENTRY_ID,
    liters: 45.0,
    total_cost: 72.50,
    price_per_l: 1.611,
    odometer: 98765.4,
    fuelled_at: "2024-06-01T08:00:00Z",
  });

  assertEquals(res.status, 201, `body: ${JSON.stringify(res.body)}`);
  const body = res.body;
  assertEquals(body.id, ENTRY_ID);
  assertEquals(body.liters, 45.0);
  assertEquals(body.total_cost, 72.50);
  assertExists(body.created_at, "created_at must be set");
  assertExists(body.updated_at, "updated_at must be set");
  assertEquals(body.deleted_at, null);

  // Store for downstream tests
  state.entryId = body.id;
  state.createdAt = body.created_at;
  state.updatedAt = body.updated_at;
});

Deno.test("POST /v1/entries — zero odometer is accepted (201)", async () => {
  const res = await client.post<FuelEntry>("/v1/entries", {
    id: `${ENTRY_ID}-zero-odo`,
    liters: 10.0,
    total_cost: 15.00,
    price_per_l: 1.500,
    odometer: 0,
    fuelled_at: "2024-06-01T08:00:00Z",
  });
  assertEquals(res.status, 201, `body: ${JSON.stringify(res.body)}`);
});

Deno.test("POST /v1/entries — missing id returns 422", async () => {
  const res = await client.post<ErrorBody>("/v1/entries", {
    liters: 45.0,
    total_cost: 72.50,
    price_per_l: 1.611,
    odometer: 98765.4,
    fuelled_at: "2024-06-01T08:00:00Z",
  });
  assertEquals(res.status, 422, `body: ${JSON.stringify(res.body)}`);
  assertExists(res.body.error, "error field must be present");
});

Deno.test("POST /v1/entries — zero liters returns 422 with message", async () => {
  const res = await client.post<ErrorBody>("/v1/entries", {
    id: `${ENTRY_ID}-bad`,
    liters: 0,
    total_cost: 72.50,
    price_per_l: 1.611,
    odometer: 98765.4,
    fuelled_at: "2024-06-01T08:00:00Z",
  });
  assertEquals(res.status, 422, `body: ${JSON.stringify(res.body)}`);
  assert(res.body.error.includes("liters"), `error should mention liters, got: ${res.body.error}`);
});

Deno.test("POST /v1/entries — negative total_cost returns 422", async () => {
  const res = await client.post<ErrorBody>("/v1/entries", {
    id: `${ENTRY_ID}-bad2`,
    liters: 45.0,
    total_cost: -1,
    price_per_l: 1.611,
    odometer: 98765.4,
    fuelled_at: "2024-06-01T08:00:00Z",
  });
  assertEquals(res.status, 422, `body: ${JSON.stringify(res.body)}`);
  assert(res.body.error.includes("total_cost"), `error should mention total_cost, got: ${res.body.error}`);
});

Deno.test("POST /v1/entries — negative odometer returns 422", async () => {
  const res = await client.post<ErrorBody>("/v1/entries", {
    id: `${ENTRY_ID}-bad3`,
    liters: 45.0,
    total_cost: 72.50,
    price_per_l: 1.611,
    odometer: -1,
    fuelled_at: "2024-06-01T08:00:00Z",
  });
  assertEquals(res.status, 422, `body: ${JSON.stringify(res.body)}`);
  assert(res.body.error.includes("odometer"), `error should mention odometer, got: ${res.body.error}`);
});

Deno.test("POST /v1/entries — missing fuelled_at returns 422", async () => {
  const res = await client.post<ErrorBody>("/v1/entries", {
    id: `${ENTRY_ID}-bad4`,
    liters: 45.0,
    total_cost: 72.50,
    price_per_l: 1.611,
    odometer: 98765.4,
  });
  assertEquals(res.status, 422, `body: ${JSON.stringify(res.body)}`);
  assertExists(res.body.error, "error field must be present");
});

// ---------- LIST ----------

Deno.test("GET /v1/entries — returns 200 with array", async () => {
  const res = await client.get<FuelEntry[]>("/v1/entries");
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(Array.isArray(res.body), "response body must be an array");
  const found = res.body.find((e) => e.id === state.entryId);
  assertExists(found, `created entry ${state.entryId} must appear in list`);
});

Deno.test("GET /v1/entries?since= — returns only entries updated after timestamp", async () => {
  // Use a timestamp far in the future so nothing qualifies
  const futureMs = Date.now() + 1_000_000_000;
  const res = await client.get<FuelEntry[]>(`/v1/entries?since=${futureMs}`);
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(Array.isArray(res.body), "response body must be an array");
  assertEquals(res.body.length, 0, "future since must return empty array");
});

Deno.test("GET /v1/entries?since= — returns entries when since is in the past", async () => {
  const pastMs = 0; // epoch — returns everything
  const res = await client.get<FuelEntry[]>(`/v1/entries?since=${pastMs}`);
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(res.body.length > 0, "since=0 must return at least the entry we created");
});

Deno.test("GET /v1/entries?since=abc — invalid since returns 400", async () => {
  const res = await client.get<ErrorBody>("/v1/entries?since=abc");
  assertEquals(res.status, 400, `body: ${JSON.stringify(res.body)}`);
  assertEquals(res.body.error, "invalid since parameter");
});

// ---------- GET BY ID ----------

Deno.test("GET /v1/entries/:id — returns the entry by id", async () => {
  const res = await client.get<FuelEntry>(`/v1/entries/${state.entryId}`);
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assertEquals(res.body.id, state.entryId);
  assertEquals(res.body.liters, 45.0);
});

Deno.test("GET /v1/entries/:id — non-existent id returns 404", async () => {
  const res = await client.get<ErrorBody>("/v1/entries/does-not-exist-xyz");
  assertEquals(res.status, 404, `body: ${JSON.stringify(res.body)}`);
  assertEquals(res.body.error, "entry not found");
});

// ---------- UPDATE ----------

Deno.test("PUT /v1/entries/:id — updates fields and returns 200", async () => {
  const newUpdatedAt = new Date(Date.now() + 5000).toISOString();

  const res = await client.put<FuelEntry>(`/v1/entries/${state.entryId}`, {
    liters: 50.0,
    total_cost: 82.00,
    price_per_l: 1.640,
    odometer: 99000.0,
    fuelled_at: "2024-06-01T08:00:00Z",
    updated_at: newUpdatedAt,
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assertEquals(res.body.liters, 50.0);
  assertEquals(res.body.total_cost, 82.00);

  state.updatedAt = res.body.updated_at;
});

Deno.test("PUT /v1/entries/:id — older updated_at is ignored (last-write-wins returns original)", async () => {
  // Send an update with a timestamp before what the server has stored
  const pastUpdatedAt = "2000-01-01T00:00:00Z";

  const res = await client.put<FuelEntry>(`/v1/entries/${state.entryId}`, {
    liters: 1.0, // should NOT take effect
    total_cost: 1.00,
    price_per_l: 1.000,
    odometer: 1.0,
    fuelled_at: "2024-06-01T08:00:00Z",
    updated_at: pastUpdatedAt,
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  // Server must return the stored entry untouched
  assertEquals(res.body.liters, 50.0, "liters must remain 50.0 — older write must be ignored");
});

// ---------- DELETE (soft) ----------

Deno.test("DELETE /v1/entries/:id — soft deletes and returns 204", async () => {
  const res = await client.delete(`/v1/entries/${state.entryId}`);
  assertEquals(res.status, 204, `body: ${JSON.stringify(res.body)}`);
});

Deno.test("GET /v1/entries/:id — deleted entry still returns with deleted_at set", async () => {
  const res = await client.get<FuelEntry>(`/v1/entries/${state.entryId}`);
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assertExists(res.body.deleted_at, "deleted_at must be set after soft delete");
  assert(res.body.deleted_at !== null, "deleted_at must not be null");
});

Deno.test("DELETE /v1/entries/:id — non-existent id returns 500", async () => {
  const res = await client.delete("/v1/entries/does-not-exist-xyz");
  assertEquals(res.status, 500, `body: ${JSON.stringify(res.body)}`);
});
