# SSO Playground

Containerised Single Sign-On playground demonstrating **OpenID Connect** authentication with
[Pocket ID](https://github.com/pocket-id/pocket-id) (passkey-based OIDC provider),
[oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy) (TLS-terminating reverse proxy with session management),
and a small **Go demo application** that consumes OIDC claims forwarded as HTTP headers.

---

## Architecture

```
                            ┌─────────────────────────┐
                            │      Pocket ID          │
                            │   (OIDC IdP, passkeys)  │
                            │   http://localhost:1411  │
                            └────────▲───────┬────────┘
                                     │       │
                          2. login   │       │ 3. auth code + redirect
                                     │       │
┌──────────┐  1. GET /   ┌───────────┴───────▼──────────┐  4. proxy   ┌──────────────┐
│  Browser ├────────────►│       oauth2-proxy            ├───────────►│   demo-app   │
│          │◄────────────┤   (TLS reverse proxy + SSO)   │◄───────────┤ (Go JSON API)│
└──────────┘  6. JSON    │   https://localhost:4443      │  5. JSON   │  :8080       │
              response   └──────────────────────────────┘            └──────────────┘
```

**Flow:**

1. User visits `https://localhost:4443`
2. oauth2-proxy has no session → redirects browser to Pocket ID for login
3. User authenticates with a passkey → Pocket ID redirects back with an auth code
4. oauth2-proxy exchanges the code for tokens, sets a session cookie, and proxies the request to the demo-app with OIDC claim headers
5. demo-app reads `X-Forwarded-User`, `X-Forwarded-Email`, `X-Forwarded-Preferred-Username`, etc.
6. Browser receives a JSON response with all decoded OIDC claims

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) (v2+)
- A browser that supports **passkeys / WebAuthn** (Chrome, Safari, Edge, Firefox 122+)

---

## Quick Start

### 1. Clone and prepare the environment file

```bash
git clone https://github.com/mkhomytsya/SSO-playground.git
cd SSO-playground

cp .env.example .env
```

### 2. Generate secrets

```bash
# Pocket ID encryption key (≥ 16 bytes) — paste into .env as ENCRYPTION_KEY
openssl rand -hex 16

# oauth2-proxy cookie secret — paste into .env as OAUTH2_PROXY_COOKIE_SECRET
openssl rand -base64 32
```

### 3. Start Pocket ID first

```bash
docker compose up -d pocket-id
```

Wait for it to become healthy:

```bash
docker compose ps   # pocket-id should show "healthy"
```

### 4. Complete Pocket ID setup

1. Open **http://localhost:1411** in your browser
2. Walk through the **setup wizard** — create your admin account with a passkey
3. Navigate to **Settings → OIDC Clients → Add OIDC Client**
4. Set the **Callback URL** to:
   ```
   https://localhost:4443/oauth2/callback
   ```
5. Copy the generated **Client ID** and **Client Secret**
6. Paste them into your `.env` file:
   ```
   OAUTH2_PROXY_CLIENT_ID=<your-client-id>
   OAUTH2_PROXY_CLIENT_SECRET=<your-client-secret>
   ```

### 5. Start all services

```bash
docker compose up -d
```

### 6. Test the SSO flow

1. Open **https://localhost:4443** (accept the self-signed certificate warning)
2. You'll be redirected to Pocket ID — authenticate with your passkey
3. After successful login, you'll see a JSON response with your OIDC claims:

```json
{
  "message": "Authenticated via OIDC (Pocket ID → oauth2-proxy)",
  "claims": {
    "user": "550e8400-e29b-41d4-a716-446655440000",
    "email": "admin@example.com",
    "preferred_username": "admin"
  },
  "all_forwarded_headers": {
    "X-Forwarded-User": "550e8400-e29b-41d4-a716-446655440000",
    "X-Forwarded-Email": "admin@example.com",
    "X-Forwarded-Preferred-Username": "admin",
    "X-Forwarded-Access-Token": "eyJ..."
  }
}
```

---

## Services

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| `cert-init` | `alpine:3.20` | — | One-shot: generates self-signed TLS certs |
| `pocket-id` | `ghcr.io/pocket-id/pocket-id` | `1411` | OIDC identity provider (passkey login) |
| `oauth2-proxy` | `quay.io/oauth2-proxy/oauth2-proxy:v7.8.1` | `4443` | TLS reverse proxy + OIDC session management |
| `demo-app` | Built from `./demo-app` | Internal | Go API returning JSON OIDC claims |

---

## File Structure

```
SSO-playground/
├── docker-compose.yml        # Full stack definition
├── .env.example              # Template — copy to .env and fill in
├── .gitignore
├── scripts/
│   └── generate-certs.sh     # Self-signed CA + server certificate
├── demo-app/
│   ├── main.go               # Go HTTP server reading claim headers
│   ├── go.mod
│   └── Dockerfile            # Multi-stage build
├── README.md
└── LICENSE
```

---

## Configuration

All configuration is done via the `.env` file. See [.env.example](.env.example) for available variables.

| Variable | Description |
|----------|-------------|
| `POCKET_ID_APP_URL` | Public URL for Pocket ID (default: `http://localhost:1411`) |
| `ENCRYPTION_KEY` | Pocket ID encryption key — at least 16 bytes (`openssl rand -hex 16`) |
| `OAUTH2_PROXY_CLIENT_ID` | OIDC Client ID from Pocket ID |
| `OAUTH2_PROXY_CLIENT_SECRET` | OIDC Client Secret from Pocket ID |
| `OAUTH2_PROXY_COOKIE_SECRET` | 32-byte base64 secret for session cookies |
| `OAUTH2_PROXY_REDIRECT_URL` | OAuth2 callback URL (default: `https://localhost:4443/oauth2/callback`) |

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| **Cookie not being set** | Ensure `OAUTH2_PROXY_COOKIE_SECURE` matches your protocol (HTTPS → `true`) |
| **OIDC discovery fails** | Check that Pocket ID is healthy: `docker compose logs pocket-id` |
| **Redirect URI mismatch** | The callback URL in Pocket ID must exactly match `OAUTH2_PROXY_REDIRECT_URL` |
| **Certificate warnings** | Expected with self-signed certs — accept in browser, or import `certs/ca.crt` |
| **Passkey not working** | Use a supported browser (Chrome/Safari/Edge). Some VM environments lack WebAuthn support |
| **"email not verified" error** | `INSECURE_OIDC_ALLOW_UNVERIFIED_EMAIL` is already set to `true` in compose |

---

## Stopping

```bash
docker compose down        # stop all services
docker compose down -v     # stop and remove volumes (reset all data)
```

---

## License

See [LICENSE](LICENSE).

