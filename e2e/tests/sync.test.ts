// ABOUTME: E2E tests for POST /v1/entries/sync — bidirectional sync with last-write-wins semantics.
// ABOUTME: Tests run sequentially; each scenario builds on server state from the previous one.

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

const SYNC_ID = `sync-e2e-${Date.now()}`;

const client = new ApiClient();

Deno.test("setup — remove leftover test rows from previous runs", async () => {
  await cleanTestData();
});

// ---------- Scenario A — Full recovery ----------

Deno.test("POST /v1/entries/sync — full recovery (last_sync_at: 0, empty entries) returns 200 array", async () => {
  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(Array.isArray(res.body), "response must be an array");
});

// ---------- Scenario B — Push new entry ----------

Deno.test("POST /v1/entries/sync — push new entry; response includes it", async () => {
  const now = new Date().toISOString();

  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [
      {
        id: SYNC_ID,
        liters: 38.0,
        total_cost: 61.00,
        price_per_l: 1.605,
        odometer: 110000,
        fuelled_at: "2024-07-01T07:30:00Z",
        updated_at: now,
      },
    ],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(Array.isArray(res.body), "response must be an array");

  const synced = res.body.find((e) => e.id === SYNC_ID);
  assertExists(synced, `entry ${SYNC_ID} must appear in sync response`);
  assertEquals(synced!.liters, 38.0);
});

// ---------- Scenario C — Last-write-wins on sync ----------

Deno.test("POST /v1/entries/sync — older updated_at is ignored by last-write-wins", async () => {
  const veryOld = "2000-01-01T00:00:00Z";

  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [
      {
        id: SYNC_ID,
        liters: 1.0, // should NOT win
        total_cost: 1.00,
        price_per_l: 1.000,
        odometer: 1,
        fuelled_at: "2024-07-01T07:30:00Z",
        updated_at: veryOld,
      },
    ],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  const entry = res.body.find((e) => e.id === SYNC_ID);
  assertExists(entry, "entry must still be present");
  assertEquals(entry!.liters, 38.0, "liters must remain 38.0 — older write must be ignored");
});

// ---------- Scenario D — Push deletion ----------

Deno.test("POST /v1/entries/sync — push deletion; entry is soft-deleted and returned", async () => {
  const deletedAt = new Date(Date.now() + 5000).toISOString();
  const updatedAt = deletedAt;

  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [
      {
        id: SYNC_ID,
        liters: 38.0,
        total_cost: 61.00,
        price_per_l: 1.605,
        odometer: 110000,
        fuelled_at: "2024-07-01T07:30:00Z",
        updated_at: updatedAt,
        deleted_at: deletedAt,
      },
    ],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  const entry = res.body.find((e) => e.id === SYNC_ID);
  assertExists(entry, "soft-deleted entry must still appear in sync response");
  assertExists(entry!.deleted_at, "deleted_at must be set on soft-deleted entry");
});

// ---------- Scenario E — Delta sync ----------

Deno.test("POST /v1/entries/sync — delta sync returns only entries updated since last_sync_at", async () => {
  // A timestamp far in the future — nothing should match
  const futureMs = Date.now() + 1_000_000_000;

  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: futureMs,
    entries: [],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(Array.isArray(res.body), "response must be an array");
  assertEquals(res.body.length, 0, "future last_sync_at must yield empty response");
});

Deno.test("POST /v1/entries/sync — delta sync since epoch returns all entries", async () => {
  const res = await client.post<FuelEntry[]>("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [],
  });

  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  assert(res.body.length > 0, "since epoch must return at least the entry we created");
});

// ---------- Edge cases ----------

Deno.test("POST /v1/entries/sync — empty body returns 400", async () => {
  const res = await client.postRaw<ErrorBody>("/v1/entries/sync");
  assertEquals(res.status, 400, `body: ${JSON.stringify(res.body)}`);
  assertEquals(res.body.error, "invalid request body");
});
