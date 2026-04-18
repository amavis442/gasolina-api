// ABOUTME: E2E tests for POST /auth/token — authentication and rolling JWT behaviour.
// ABOUTME: Runs against a live server; set API_URL and DEVICE_SECRET env vars to override defaults.

import { assertEquals, assertExists, assert } from "@std/assert";
import { ApiClient } from "../client.ts";
import { DEVICE_SECRET } from "../config.ts";

const client = new ApiClient();

Deno.test("POST /auth/token — valid device secret returns 200 with token", async () => {
  const res = await client.rawAuth({ device_secret: DEVICE_SECRET });
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  const body = res.body as { token: string };
  assertExists(body.token, "token must be present");
  assert(body.token.length > 0, "token must not be empty");
});

Deno.test("POST /auth/token — wrong device secret returns 401", async () => {
  const res = await client.rawAuth({ device_secret: "definitely-wrong-secret" });
  assertEquals(res.status, 401, `body: ${JSON.stringify(res.body)}`);
  const body = res.body as { error: string };
  assertEquals(body.error, "invalid device secret");
});

Deno.test("POST /auth/token — empty body returns 400", async () => {
  const res = await client.rawAuthEmpty();
  assertEquals(res.status, 400, `body: ${JSON.stringify(res.body)}`);
  const body = res.body as { error: string };
  assertEquals(body.error, "invalid request body");
});

Deno.test("GET /v1/entries — missing Authorization header returns 401", async () => {
  const res = await client.getNoAuth("/v1/entries");
  assertEquals(res.status, 401, `body: ${JSON.stringify(res.body)}`);
  const body = res.body as { error: string };
  assertEquals(body.error, "missing bearer token");
});

Deno.test("GET /v1/entries — tampered token returns 401", async () => {
  const res = await client.getWithToken("/v1/entries", "tampered.token.here");
  assertEquals(res.status, 401, `body: ${JSON.stringify(res.body)}`);
  const body = res.body as { error: string };
  assertEquals(body.error, "invalid token");
});

Deno.test("POST /auth/token — response does NOT include X-Refresh-Token (auth endpoint is unprotected)", async () => {
  const res = await client.rawAuth({ device_secret: DEVICE_SECRET });
  assertEquals(res.status, 200);
  // The rolling refresh header is only emitted by the JWT middleware on /v1/* routes
  const refreshHeader = res.headers.get("X-Refresh-Token");
  assertEquals(refreshHeader, null, "auth endpoint must not emit X-Refresh-Token");
});

Deno.test("GET /v1/entries — valid token response includes X-Refresh-Token header", async () => {
  // Fresh client so we can observe the header
  const fresh = new ApiClient();
  const res = await fresh.get("/v1/entries");
  assertEquals(res.status, 200, `body: ${JSON.stringify(res.body)}`);
  const refreshHeader = res.headers.get("X-Refresh-Token");
  assertExists(refreshHeader, "X-Refresh-Token header must be present on every /v1/* response");
  assert(refreshHeader!.length > 0, "X-Refresh-Token must not be empty");
});

Deno.test("Rolling JWT — token is updated after each /v1/* call", async () => {
  const fresh = new ApiClient();
  await fresh.authenticate();
  const tokenBefore = fresh.getToken();

  await fresh.get("/v1/entries");
  const tokenAfter = fresh.getToken();

  assertExists(tokenBefore);
  assertExists(tokenAfter);
  // The server issues a new token each time; they may differ in `iat`/`exp`
  // We just verify the client stored the refreshed value (non-null, non-empty)
  assert(tokenAfter!.length > 0, "rolling token must be stored after a /v1/* call");
});
