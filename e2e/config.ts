// ABOUTME: Runtime configuration for the E2E test suite.
// ABOUTME: Reads directly from config.json at the repo root — no environment variables.

interface AppConfig {
  database_url: string;
  jwt_secret: string;
  device_secret: string;
  port: string;
  token_ttl: string;
}

const raw = await Deno.readTextFile(new URL("../config.json", import.meta.url));
const cfg: AppConfig = JSON.parse(raw);

export const BASE_URL = `http://localhost:${cfg.port}`;
export const DEVICE_SECRET = cfg.device_secret;
