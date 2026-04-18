# Gasolina API — Postman Test Plan

## 1. Environment setup

Create a Postman **Environment** named `Gasolina Local` with these variables:

| Variable       | Initial value              | Notes |
|---|---|---|
| `base_url`     | `http://localhost:8080`    | Change for staging/prod |
| `device_secret`| *(value from config.json)* | |
| `token`        | *(leave empty)*            | Auto-filled by the Auth test |

Set this environment as active before running any request.

---

## 2. Rolling JWT — auto-refresh script

Every `/v1/*` response contains an `X-Refresh-Token` header with a new token.
Add the script below to the **Postman Collection → Tests** tab so it runs after every request automatically:

```javascript
const refreshed = pm.response.headers.get("X-Refresh-Token");
if (refreshed) {
    pm.environment.set("token", refreshed);
}
```

This means you only need to call `/auth/token` once per session.

---

## 3. Test cases

### 3.1 POST /auth/token

**Purpose:** Exchange device secret for a JWT.

**Request**
- Method: `POST`
- URL: `{{base_url}}/auth/token`
- Body (JSON):
```json
{ "device_secret": "{{device_secret}}" }
```

**Tests tab** (add this only on the auth request to seed the token):
```javascript
pm.test("Status is 200", () => pm.response.to.have.status(200));

pm.test("Response contains token", () => {
    const body = pm.response.json();
    pm.expect(body.token).to.be.a("string").and.not.empty;
    pm.environment.set("token", body.token);
});
```

**Expected response (200):**
```json
{ "token": "eyJhbGci..." }
```

---

#### Edge cases

| Scenario | Body | Expected |
|---|---|---|
| Wrong device secret | `{"device_secret": "wrong"}` | `401` + `{"error": "invalid device secret"}` |
| Empty body | *(no body)* | `400` + `{"error": "invalid request body"}` |
| Missing `Authorization` on `/v1/*` | *(omit header)* | `401` + `{"error": "missing bearer token"}` |
| Tampered token on `/v1/*` | `Bearer tampered.token.here` | `401` + `{"error": "invalid token"}` |

---

### 3.2 POST /v1/entries — Create

**Purpose:** Push a new fuel entry.

**Request**
- Method: `POST`
- URL: `{{base_url}}/v1/entries`
- Headers: `Authorization: Bearer {{token}}`
- Body (JSON):
```json
{
  "id":         "entry-001",
  "liters":     45.0,
  "total_cost": 72.50,
  "price_per_l": 1.611,
  "odometer":   98765.4,
  "fuelled_at": "2024-06-01T08:00:00Z"
}
```

**Tests tab:**
```javascript
pm.test("Status is 201", () => pm.response.to.have.status(201));

pm.test("Entry is returned", () => {
    const body = pm.response.json();
    pm.expect(body.id).to.eql("entry-001");
    pm.expect(body.created_at).to.be.a("string").and.not.empty;
    pm.expect(body.updated_at).to.be.a("string").and.not.empty;
});

// Save id for later requests
pm.environment.set("entry_id", pm.response.json().id);
```

**Expected response (201):** full FuelEntry object including `created_at` and `updated_at`.

---

#### Edge cases

| Scenario | Change to body | Expected |
|---|---|---|
| Missing `id` | Remove `id` field | `422` + validation error |
| Zero liters | `"liters": 0` | `422` + `"liters must be positive"` |
| Negative total_cost | `"total_cost": -1` | `422` + `"total_cost must be positive"` |
| Negative odometer | `"odometer": -1` | `422` + `"odometer must be non-negative"` |
| Zero odometer | `"odometer": 0` | `201` — zero odometer is valid |
| Missing fuelled_at | Remove `fuelled_at` | `422` + `"fuelled_at is required"` |

---

### 3.3 GET /v1/entries — List all

**Purpose:** Pull all entries.

**Request**
- Method: `GET`
- URL: `{{base_url}}/v1/entries`
- Headers: `Authorization: Bearer {{token}}`

**Tests tab:**
```javascript
pm.test("Status is 200", () => pm.response.to.have.status(200));

pm.test("Response is an array", () => {
    pm.expect(pm.response.json()).to.be.an("array");
});
```

---

#### Edge cases

| Scenario | URL | Expected |
|---|---|---|
| Delta pull | `{{base_url}}/v1/entries?since=1717200000000` | `200` — only entries with `updated_at` after the timestamp |
| Invalid since param | `{{base_url}}/v1/entries?since=abc` | `400` + `{"error": "invalid since parameter"}` |
| Full recovery (all) | `{{base_url}}/v1/entries` (no since) | `200` — all entries |

> Tip: `since` is Unix time in **milliseconds**. Get the current ms timestamp in Postman pre-request script: `pm.environment.set("now_ms", Date.now());`

---

### 3.4 GET /v1/entries/:id — Get by ID

**Purpose:** Pull a single entry.

**Request**
- Method: `GET`
- URL: `{{base_url}}/v1/entries/{{entry_id}}`
- Headers: `Authorization: Bearer {{token}}`

**Tests tab:**
```javascript
pm.test("Status is 200", () => pm.response.to.have.status(200));

pm.test("Correct entry returned", () => {
    pm.expect(pm.response.json().id).to.eql(pm.environment.get("entry_id"));
});
```

---

#### Edge cases

| Scenario | URL | Expected |
|---|---|---|
| Non-existent ID | `{{base_url}}/v1/entries/does-not-exist` | `404` + `{"error": "entry not found"}` |

---

### 3.5 PUT /v1/entries/:id — Update

**Purpose:** Push an update. Last-write-wins on `updated_at`.

**Request**
- Method: `PUT`
- URL: `{{base_url}}/v1/entries/{{entry_id}}`
- Headers: `Authorization: Bearer {{token}}`
- Body (JSON):
```json
{
  "liters":     50.0,
  "total_cost": 82.00,
  "price_per_l": 1.640,
  "odometer":   99000.0,
  "fuelled_at": "2024-06-01T08:00:00Z",
  "updated_at": "2024-06-01T10:00:00Z"
}
```

**Tests tab:**
```javascript
pm.test("Status is 200", () => pm.response.to.have.status(200));

pm.test("Liters updated", () => {
    pm.expect(pm.response.json().liters).to.eql(50.0);
});
```

---

#### Edge cases

| Scenario | Change | Expected |
|---|---|---|
| `updated_at` older than stored | Set `updated_at` to a past time | `200` — body returns **original** unchanged entry (last-write-wins ignores older writes) |
| Non-existent ID | Change URL id | `422` + error |

---

### 3.6 DELETE /v1/entries/:id — Soft delete

**Purpose:** Mark an entry as deleted (sets `deleted_at`; entry is not removed from the DB).

**Request**
- Method: `DELETE`
- URL: `{{base_url}}/v1/entries/{{entry_id}}`
- Headers: `Authorization: Bearer {{token}}`

**Tests tab:**
```javascript
pm.test("Status is 204", () => pm.response.to.have.status(204));
```

Verify the soft-delete by calling `GET /v1/entries/{{entry_id}}` — the entry still returns but now has a `deleted_at` value.

---

#### Edge cases

| Scenario | Expected |
|---|---|
| Non-existent ID | `500` + error |

---

### 3.7 POST /v1/entries/sync — Bidirectional sync

**Purpose:** Merge client entries with the server. The server applies last-write-wins and returns all entries updated since `last_sync_at`.

#### Scenario A — Full recovery (`last_sync_at: 0`)

**Request**
- Method: `POST`
- URL: `{{base_url}}/v1/entries/sync`
- Headers: `Authorization: Bearer {{token}}`
- Body (JSON):
```json
{
  "last_sync_at": 0,
  "entries": []
}
```

**Expected:** `200` — array of **all** entries on the server.

---

#### Scenario B — Push new entries

```json
{
  "last_sync_at": 0,
  "entries": [
    {
      "id":         "entry-sync-001",
      "liters":     38.0,
      "total_cost": 61.00,
      "price_per_l": 1.605,
      "odometer":   110000,
      "fuelled_at": "2024-07-01T07:30:00Z",
      "updated_at": "2024-07-01T07:30:00Z"
    }
  ]
}
```

**Expected:** `200` — array includes the newly synced entry.

---

#### Scenario C — Push a deletion

```json
{
  "last_sync_at": 0,
  "entries": [
    {
      "id":         "entry-sync-001",
      "liters":     38.0,
      "total_cost": 61.00,
      "price_per_l": 1.605,
      "odometer":   110000,
      "fuelled_at": "2024-07-01T07:30:00Z",
      "updated_at": "2024-07-02T09:00:00Z",
      "deleted_at": "2024-07-02T09:00:00Z"
    }
  ]
}
```

**Expected:** `200` — entry is still returned but has `deleted_at` set.

---

#### Scenario D — Delta sync

```json
{
  "last_sync_at": 1719820800000,
  "entries": []
}
```

**Expected:** `200` — only entries with `updated_at` after the given Unix ms timestamp.

---

#### Edge cases

| Scenario | Expected |
|---|---|
| Empty body (no JSON) | `400` + `{"error": "invalid request body"}` |

---

## 4. Suggested run order

Run requests in this order for a complete happy-path flow:

1. `POST /auth/token` — get token
2. `POST /v1/entries` — create entry-001
3. `GET /v1/entries` — confirm it appears
4. `GET /v1/entries/{{entry_id}}` — fetch by id
5. `PUT /v1/entries/{{entry_id}}` — update liters
6. `GET /v1/entries/{{entry_id}}` — verify update
7. `POST /v1/entries/sync` (full recovery) — confirm sync returns all
8. `DELETE /v1/entries/{{entry_id}}` — soft delete
9. `GET /v1/entries/{{entry_id}}` — verify deleted_at is set
10. `POST /v1/entries/sync` (delta) — confirm deleted entry propagates

---

## 5. Verifying the rolling token

After any `/v1/*` response, check the Postman console — the `X-Refresh-Token` header should be present and the `token` environment variable should update automatically (via the collection-level test script from section 2).
