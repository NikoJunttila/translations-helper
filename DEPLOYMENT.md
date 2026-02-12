# Deployment Guide

This guide covers how to pull and run the JSON Translation Editor as a container using Docker or Podman.

## Pulling the Image

The container image is automatically built and pushed to GitHub Container Registry on every push to `master` and on version tags. It supports both **x86-64 (amd64)** and **ARM64** (e.g., Raspberry Pi, Apple Silicon).

```bash
# Latest from master
docker pull ghcr.io/nikojunttila/translations-helper:latest

# Specific commit
docker pull ghcr.io/nikojunttila/translations-helper:<commit-sha>

# Specific version (when a tag like v1.0.0 is pushed)
docker pull ghcr.io/nikojunttila/translations-helper:1.0.0
```

> [!NOTE]
> Replace `docker` with `podman` in any command below if you prefer Podman.

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ Yes | — | Database connection string (see [Database Setup](#database-setup)) |
| `DATABASE_AUTH_TOKEN` | Only for Turso | — | Auth token for Turso cloud database |
| `PORT` | No | `8090` | HTTP server port |
| `METRICS_PORT` | ✅ Yes | — | Port for Prometheus metrics endpoint |
| `GO_ENV` | No | `production` | Set automatically in the container |
| `OPENAI_API_KEY` | No | — | Required for the AI auto-translate feature |
| `OTLP_ENDPOINT` | No | — | OpenTelemetry Collector gRPC endpoint for tracing |

---

## Getting API Keys

### OpenAI API Key (for Auto-Translate)

The auto-translate feature uses OpenAI's `gpt-4o` model to translate missing fields automatically.

1. Go to [platform.openai.com](https://platform.openai.com/) and sign up or log in
2. Navigate to **API Keys** in the left sidebar (or go to [platform.openai.com/api-keys](https://platform.openai.com/api-keys))
3. Click **Create new secret key**, give it a name, and copy the key
4. Add billing/credits under **Settings → Billing** — the API is pay-per-use

> [!TIP]
> Only missing translation fields are sent to the API, so costs are minimal for typical usage.

### Turso Database (optional — cloud database)

You can use a local SQLite file instead of Turso. If you want a cloud database:

1. Install the Turso CLI: `curl -sSfL https://get.tur.so/install.sh | bash`
2. Sign up: `turso auth signup`
3. Create a database: `turso db create translations`
4. Get the connection URL: `turso db show translations --url`
5. Create an auth token: `turso db tokens create translations`

Set `DATABASE_URL` to the URL from step 4 and `DATABASE_AUTH_TOKEN` to the token from step 5.

---

## Database Setup

The application supports two database backends:

### Local SQLite (simplest)
```bash
DATABASE_URL=file:./data/local.db
```
The database file and tables are created automatically on first run. Mount a volume to persist data (see below).

### Turso / libSQL (cloud)
```bash
DATABASE_URL=libsql://your-db-name.turso.io
DATABASE_AUTH_TOKEN=your-token-here
```

---

## Running the Container

### Docker

```bash
docker run -d \
  --name translations-helper \
  -p 8090:8090 \
  -p 8081:8081 \
  -v translations-data:/app/data \
  -e DATABASE_URL=file:./data/local.db \
  -e METRICS_PORT=8081 \
  -e OPENAI_API_KEY=sk-... \
  ghcr.io/nikojunttila/translations-helper:latest
```

### Podman

```bash
podman run -d \
  --name translations-helper \
  -p 8090:8090 \
  -p 8081:8081 \
  -v translations-data:/app/data:Z \
  -e DATABASE_URL=file:./data/local.db \
  -e METRICS_PORT=8081 \
  -e OPENAI_API_KEY=sk-... \
  ghcr.io/nikojunttila/translations-helper:latest
```

### Using an env file

Create a file called `translations.env`:
```env
DATABASE_URL=file:./data/local.db
METRICS_PORT=8081
OPENAI_API_KEY=sk-...
# OTLP_ENDPOINT=your-collector:4317
```

Then run with:
```bash
docker run -d \
  --name translations-helper \
  -p 8090:8090 \
  -p 8081:8081 \
  -v translations-data:/app/data \
  --env-file translations.env \
  ghcr.io/nikojunttila/translations-helper:latest
```

---

## Data Persistence

When using local SQLite, mount a volume to `/app/data` so your database survives container restarts:

```bash
# Named volume (recommended)
-v translations-data:/app/data

# Bind mount to host directory
-v /path/on/host:/app/data
```

---

## Observability

### Prometheus Metrics

The app exposes a Prometheus-compatible metrics endpoint on the `METRICS_PORT`. Add a scrape target to your Prometheus config:

```yaml
scrape_configs:
  - job_name: "translations-helper"
    static_configs:
      - targets: ["<container-host>:8081"]
```

### OpenTelemetry Tracing

Set `OTLP_ENDPOINT` to your OpenTelemetry Collector's gRPC endpoint to enable distributed tracing:

```bash
-e OTLP_ENDPOINT=your-collector:4317
```

The application sends traces via gRPC using the OTLP exporter. Compatible with any OTLP-capable backend (Grafana Tempo, Jaeger, etc.).

---

## Ports

| Port | Purpose |
|---|---|
| `8090` | Main HTTP server (UI + API) |
| `8081` | Prometheus metrics (configurable via `METRICS_PORT`) |
