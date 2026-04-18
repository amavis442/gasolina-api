// ABOUTME: Database cleanup helper for the E2E test suite.
// ABOUTME: Deletes all rows with test-prefixed IDs before each suite runs.

import pg from "npm:pg";
import { BASE_URL } from "./config.ts";

const raw = await Deno.readTextFile(new URL("../config.json", import.meta.url));
const cfg = JSON.parse(raw);

export async function cleanTestData(): Promise<void> {
  const client = new pg.Client({ connectionString: cfg.database_url });
  await client.connect();
  try {
    await client.query(
      `DELETE FROM fuel_entries WHERE id LIKE 'e2e-%' OR id LIKE 'sync-e2e-%'`
    );
  } finally {
    await client.end();
  }
}

export { BASE_URL };
