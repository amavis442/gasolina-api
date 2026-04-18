// ABOUTME: HTTP client for the Gasolina API E2E tests.
// ABOUTME: Handles Bearer token injection and rolling X-Refresh-Token header automatically.

import { BASE_URL, DEVICE_SECRET } from "./config.ts";

export interface ApiResponse<T = unknown> {
  status: number;
  body: T;
  headers: Headers;
}

export class ApiClient {
  private token: string | null = null;

  constructor(
    private baseUrl: string = BASE_URL,
    private deviceSecret: string = DEVICE_SECRET,
  ) {}

  // ---------- token management ----------

  async authenticate(): Promise<ApiResponse<{ token: string }>> {
    const res = await fetch(`${this.baseUrl}/auth/token`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ device_secret: this.deviceSecret }),
    });
    const body = await res.json();
    if (res.ok) this.token = body.token;
    return { status: res.status, body, headers: res.headers };
  }

  private async ensureToken(): Promise<void> {
    if (!this.token) await this.authenticate();
  }

  private captureRefresh(res: Response): void {
    const refreshed = res.headers.get("X-Refresh-Token");
    if (refreshed) this.token = refreshed;
  }

  getToken(): string | null {
    return this.token;
  }

  setToken(token: string): void {
    this.token = token;
  }

  // ---------- raw auth (no automatic token) ----------

  async rawAuth(body: unknown): Promise<ApiResponse> {
    const res = await fetch(`${this.baseUrl}/auth/token`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return { status: res.status, body: await res.json(), headers: res.headers };
  }

  async rawAuthEmpty(): Promise<ApiResponse> {
    const res = await fetch(`${this.baseUrl}/auth/token`, {
      method: "POST",
    });
    const text = await res.text();
    let body: unknown;
    try {
      body = JSON.parse(text);
    } catch {
      body = text;
    }
    return { status: res.status, body, headers: res.headers };
  }

  // ---------- authenticated helpers ----------

  async get<T = unknown>(path: string): Promise<ApiResponse<T>> {
    await this.ensureToken();
    const res = await fetch(`${this.baseUrl}${path}`, {
      headers: { Authorization: `Bearer ${this.token}` },
    });
    this.captureRefresh(res);
    return { status: res.status, body: await res.json() as T, headers: res.headers };
  }

  async post<T = unknown>(path: string, body: unknown): Promise<ApiResponse<T>> {
    await this.ensureToken();
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify(body),
    });
    this.captureRefresh(res);
    const text = await res.text();
    let parsed: unknown;
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = text;
    }
    return { status: res.status, body: parsed as T, headers: res.headers };
  }

  async postRaw<T = unknown>(path: string): Promise<ApiResponse<T>> {
    await this.ensureToken();
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${this.token}` },
    });
    this.captureRefresh(res);
    return { status: res.status, body: await res.json() as T, headers: res.headers };
  }

  async put<T = unknown>(path: string, body: unknown): Promise<ApiResponse<T>> {
    await this.ensureToken();
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify(body),
    });
    this.captureRefresh(res);
    return { status: res.status, body: await res.json() as T, headers: res.headers };
  }

  async delete(path: string): Promise<ApiResponse<unknown>> {
    await this.ensureToken();
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${this.token}` },
    });
    this.captureRefresh(res);
    const text = await res.text();
    let body: unknown = null;
    if (text) {
      try {
        body = JSON.parse(text);
      } catch {
        body = text;
      }
    }
    return { status: res.status, body, headers: res.headers };
  }

  // ---------- unauthenticated helpers for negative tests ----------

  async getNoAuth<T = unknown>(path: string): Promise<ApiResponse<T>> {
    const res = await fetch(`${this.baseUrl}${path}`);
    return { status: res.status, body: await res.json() as T, headers: res.headers };
  }

  async getWithToken<T = unknown>(path: string, token: string): Promise<ApiResponse<T>> {
    const res = await fetch(`${this.baseUrl}${path}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    return { status: res.status, body: await res.json() as T, headers: res.headers };
  }
}
