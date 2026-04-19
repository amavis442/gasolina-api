// ABOUTME: Quick smoke test that verifies the live API is reachable and behaving correctly
// ABOUTME: Usage: deno run --allow-net --allow-read tools/smoke_test.ts [base_url]

import { ApiClient } from "../e2e/client.ts";
import { DEVICE_SECRET } from "../e2e/config.ts";

const baseUrl = Deno.args[0]?.replace(/\/$/, "") ?? "http://localhost:8080";

console.log(`\nSmoke testing: ${baseUrl}\n${"─".repeat(50)}`);

const client = new ApiClient(baseUrl, DEVICE_SECRET);
let passed = 0;
let failed = 0;

async function check(label: string, fn: () => Promise<void>) {
  try {
    await fn();
    console.log(`  ✓  ${label}`);
    passed++;
  } catch (e) {
    console.log(`  ✗  ${label}`);
    console.log(`     ${e}`);
    failed++;
  }
}

function assert(cond: boolean, msg: string) {
  if (!cond) throw new Error(msg);
}

// ── Auth ──────────────────────────────────────────────────────────────────���───

await check("POST /auth/token  — valid secret returns 200 with token", async () => {
  const res = await client.authenticate();
  assert(res.status === 200, `expected 200, got ${res.status}`);
  const body = res.body as { token: string };
  assert(typeof body.token === "string" && body.token.length > 0, "no token in response");
});

await check("POST /auth/token  — wrong secret returns 401", async () => {
  const bad = new ApiClient(baseUrl, "wrong-secret");
  const res = await bad.rawAuth({ device_secret: "wrong-secret" });
  assert(res.status === 401, `expected 401, got ${res.status}`);
});

await check("POST /auth/token  — empty body returns 400", async () => {
  const bad = new ApiClient(baseUrl, DEVICE_SECRET);
  const res = await bad.rawAuthEmpty();
  assert(res.status === 400, `expected 400, got ${res.status}`);
});

// ── Auth middleware ────────────────────────────────────────────────────────────

await check("GET  /v1/entries  — missing token returns 401", async () => {
  const unauth = new ApiClient(baseUrl, DEVICE_SECRET);
  const res = await unauth.getNoAuth("/v1/entries");
  assert(res.status === 401, `expected 401, got ${res.status}`);
});

await check("GET  /v1/entries  — tampered token returns 401", async () => {
  const unauth = new ApiClient(baseUrl, DEVICE_SECRET);
  const res = await unauth.getWithToken("/v1/entries", "bad.token.here");
  assert(res.status === 401, `expected 401, got ${res.status}`);
});

await check("GET  /v1/entries  — valid token returns 200 + X-Refresh-Token", async () => {
  const res = await client.get("/v1/entries");
  assert(res.status === 200, `expected 200, got ${res.status}`);
  const refresh = res.headers.get("x-refresh-token");
  assert(!!refresh && refresh.length > 0, "missing X-Refresh-Token header");
});

// ── Entries CRUD ──────────────────────────────────────────────────────────────

const testId = `smoke-${Date.now()}`;

await check("POST /v1/entries  — create entry returns 201", async () => {
  const res = await client.post("/v1/entries", {
    id: testId,
    liters: 42.0,
    total_cost: 68.50,
    price_per_l: 1.631,
    kilometers: 520.0,
    fuelled_at: new Date().toISOString(),
  });
  assert(res.status === 201, `expected 201, got ${res.status}: ${JSON.stringify(res.body)}`);
  const body = res.body as { id: string };
  assert(body.id === testId, `expected id=${testId}, got ${body.id}`);
});

await check("GET  /v1/entries/{id} — fetch created entry returns 200", async () => {
  const res = await client.get(`/v1/entries/${testId}`);
  assert(res.status === 200, `expected 200, got ${res.status}`);
  const body = res.body as { kilometers: number };
  assert(body.kilometers === 520.0, `expected kilometers=520, got ${body.kilometers}`);
});

await check("PUT  /v1/entries/{id} — update entry returns 200", async () => {
  const res = await client.put(`/v1/entries/${testId}`, {
    liters: 44.0,
    total_cost: 71.00,
    price_per_l: 1.614,
    kilometers: 540.0,
    fuelled_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  });
  assert(res.status === 200, `expected 200, got ${res.status}: ${JSON.stringify(res.body)}`);
});

await check("POST /v1/entries/sync — bulk sync returns 200", async () => {
  const res = await client.post("/v1/entries/sync", {
    last_sync_at: 0,
    entries: [],
  });
  assert(res.status === 200, `expected 200, got ${res.status}`);
  assert(Array.isArray(res.body), "expected array response");
});

await check("DELETE /v1/entries/{id} — soft-delete returns 204", async () => {
  const res = await client.delete(`/v1/entries/${testId}`);
  assert(res.status === 204, `expected 204, got ${res.status}`);
});

// ── Summary ───────────────────────────────────────────────────────────────────

console.log(`\n${"─".repeat(50)}`);
console.log(`  ${passed} passed  ${failed > 0 ? failed + " failed" : ""}`);
if (failed > 0) Deno.exit(1);
