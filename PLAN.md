# AmneziaWG Syncloud App — Implementation Plan

## Goal

Build a Syncloud app that runs an **AmneziaWG** VPN server on a Syncloud device, so users behind hostile DPI (Russia/TSPU, Iran, China) can reach their home network when OpenVPN gets blocked.

Packaged as a snap with the same shape as the sibling `openvpn` app (`../openvpn`), but running AmneziaWG instead of OpenVPN, and with a web UI to generate client configs / QR codes / manage peers.

This is a **new, separate app** — not a mode inside the openvpn app, and not the same as a future plain `wireguard` app. See "Why separate app" below.

---

## Why this exists (context for the implementer)

- Existing `openvpn` Syncloud app works fine globally, but fails for users behind TSPU (Russian state DPI). Observed behaviour: OpenVPN handshake completes, then data packets are silently dropped — the client reconnects every 120–240 s, tunnel is useless.
- `zapret` running on the client partially helps for HTTPS but interferes with OpenVPN UDP and makes things worse.
- OpenVPN over TCP/443 + `tls-crypt-v2` is one mitigation, but AmneziaWG is purpose-built for this threat model and is the more robust answer.
- AmneziaWG is a fork of WireGuard with obfuscation (junk packets, random padding, customizable header magic bytes) that breaks DPI fingerprinting.

## Why separate app (not merged with wireguard)

- A future plain `wireguard` Syncloud app is desirable: maximum performance, stock clients on every OS, no obfuscation overhead.
- AmneziaWG clients need the Amnezia app (or `awg`/`awg-quick`) — not interchangeable with stock WG clients for users who only have built-in tooling.
- Config surface differs: AmneziaWG adds `Jc / Jmin / Jmax / S1 / S2 / H1..H4` knobs that are nonsensical for plain WG.
- Kernel module conflict: AmneziaWG ships its own module (`amneziawg`), plain WG uses mainline.
- Threat model is explicit when the app is named for what it does.

Shared internal logic (key gen, peer management) can be factored into a small Go library that both apps consume — but keep the snaps distinct.

---

## Reference apps

`../openvpn` is the closest *functional* sibling (VPN server + web UI), but its **architecture is outdated** — Python installer, Go/Beego HTML templates, `meta/snap.yaml`, cookie forward-auth. Use it only for the VPN-specific bits: IPv6 prefix-delegation (`bin/prefix_delegation.sh`), server-config template structure.

For the *architecture*, mirror these newer siblings:

- **`../owncast`** — canonical modern shape: `snap.yaml` at repo root, Go + Cobra CLI (`cli/cmd/cli/main.go`), Unix sockets for internal comms, statically-built vendored upstream (`backend/build.sh` uses `-ldflags "-linkmode external -extldflags -static"`). Two daemons: `backend` + `nginx`.
- **`../owntracks`** — same shape as owncast. Good second reference for Unix-socket nginx config (`config/nginx.conf`).
- **`../paperless`** — the **OIDC auth pattern**. `platformClient.RegisterOIDCClient(...)` call in `paperless/cli/installer/installer.go:275`. No other sibling app currently uses OIDC; this is the one reference.
- **`../redirect`** — the **Playwright UI-test pattern**. `www/e2e/*.spec.js`, `www/playwright.config.js`, CI runs in `mcr.microsoft.com/playwright:v1.59.1-jammy` via `ci/ui.sh desktop`. Also confirms **Vue 3 + Element Plus + Vite** as the app frontend stack (matches `../platform/web/platform/`).

### Existing `openvpn` app layout

Location: `../openvpn/` (i.e. `/data/data/com.termux/files/home/storage/projects/openvpn`).

```
openvpn/
├── LICENSE
├── README.md
├── package.sh              # builds the snap via mksquashfs
├── meta/
│   ├── snap.yaml           # snap apps: web, server, nginx + confinement
│   └── hooks/              # install/configure/post-refresh hook scripts
├── hooks/
│   ├── installer.py        # Python installer using syncloudlib
│   ├── storage-change
│   └── storage-change.py
├── bin/
│   ├── generate-keys.sh
│   ├── prefix_delegation.sh
│   ├── service.server.sh   # starts openvpn --config server.conf
│   ├── service.web.sh      # starts the Go web UI
│   └── service.nginx.sh    # starts nginx fronting the web UI
├── config/                 # templated at install time
│   ├── nginx.conf
│   ├── easyrsa/
│   ├── pki/                # ca, index.txt, serial
│   └── web/app.conf
├── openvpn/
│   ├── bin/                # upstream openvpn binaries
│   ├── build.sh
│   └── test.sh
├── easyrsa/                # easyrsa scripts for cert/key gen
├── openssl/                # bundled openssl
├── nginx/                  # bundled nginx
├── python/                 # bundled python runtime used by installer
├── web/
│   ├── server/             # Go web server
│   │   ├── config/templates/server-config.tpl
│   │   └── mi/             # management-interface client
│   ├── controllers/        # Go controllers: login, certificates, logs, settings, ovconfig, profile, api-*
│   ├── conf/
│   ├── static/
│   └── views/              # HTML templates
└── test/
```

The `server-config.tpl` is the OpenVPN server config template. The web UI is a Go/Beego app (imports `github.com/adamwalach/openvpn-web-ui` controllers) that talks to OpenVPN's management interface via a Unix socket.

### Snap layout at runtime (on the Syncloud device)

- `/snap/openvpn/current/` — read-only app code (binaries, templates)
- `/var/snap/openvpn/current/` — writable data: `config/`, `openvpn/server.conf`, `pki/`, sockets
- `/var/snap/openvpn/common/` — persistent data: `data.db`, `db/`, `installed`, `log/`, `nginx/`, sockets

systemd units: `snap.openvpn.server.service`, `snap.openvpn.web.service`, `snap.openvpn.nginx.service` — all three `active running`.

---

## Target design for `amneziawg` app

### Runtime components

1. **awg server** — the AmneziaWG data plane. Either:
   - kernel module `amneziawg` + `awg-tools` (`awg`, `awg-quick`), or
   - userspace `amneziawg-go` (easier inside snap confinement; no kernel build).

   Start with userspace `amneziawg-go` bundled in the snap; revisit kernel module later if throughput is insufficient.

2. **web UI** — two pieces:
   - **Backend**: Go HTTP API (no HTML templates). Handles OIDC callback, session cookie, peer CRUD, serves the SPA static bundle.
   - **Frontend**: **Vue 3 + Element Plus + Vite + TypeScript** SPA. Mirror the toolchain of `../platform/web/platform/` (same `vite.config.js` shape, Element Plus auto-import, Pinia for state, Vue Router).

   Pages:
   - **Dashboard**: status, current peers, RX/TX bytes per peer, last handshake time. Also displays **"forward UDP port N on your router"** reminder (single port for the whole app — all clients share it).
   - **Peers**: add/remove peers, generate config file + QR code, revoke.
   - **Server settings**: port, interface address pool, MTU, obfuscation params (`Jc`, `Jmin`, `Jmax`, `S1`, `S2`, `H1`, `H2`, `H3`, `H4`).
   - **Logs**: tail systemd journal for the app.
   - **Login** / auth: **OIDC via the Syncloud platform** (Authelia). See "Auth" section below.

3. **nginx** fronts the web UI over Unix socket (not TCP). Platform router hits `/var/snap/amneziawg/common/web.socket`; nginx forwards `/api/*` to the Go backend's own socket at `/var/snap/amneziawg/current/backend.sock`. Same pattern as `../owncast/config/nginx.conf` and `../owntracks/config/nginx.conf`.

### Internal communication — Unix sockets only

One external port only: the UDP VPN listener. Everything else is Unix-socket:

| Component          | Socket path                                              |
|--------------------|----------------------------------------------------------|
| web (platform → nginx) | `/var/snap/amneziawg/common/web.socket`              |
| backend (nginx → Go)   | `/var/snap/amneziawg/current/backend.sock`           |

No localhost TCP. No HTTPS on the app itself — the platform's reverse proxy terminates TLS.

### Snap services (mirror of `../owncast/snap.yaml` — root-level, not `meta/`)

```yaml
apps:
  backend:
    command: bin/service.backend.sh
    daemon: simple
    restart-condition: always
    start-timeout: 600s
    restart-delay: 10s
  server:
    command: bin/service.server.sh
    daemon: simple
    restart-condition: always
    start-timeout: 600s
    restart-delay: 10s
  nginx:
    command: bin/service.nginx.sh
    daemon: simple
    restart-condition: always
    start-timeout: 600s
    restart-delay: 10s

description: AmneziaWG VPN server (WireGuard with DPI obfuscation)
grade: stable
name: amneziawg
summary: AmneziaWG
```

**No `confinement:` line, no `plugs:` lists.** Syncloud ships a custom snapd fork; confinement isn't enforced, so we drop the OpenVPN-era strict/classic + `network-control` discussion entirely.

### Installer (Go, mirror `../paperless/cli/installer/`)

**Go installer, not Python.** The openvpn app uses Python + `syncloudlib`, which is the old pattern. Paperless uses Go + `github.com/syncloud/golib/platform`, which is cleaner and what new apps should follow.

Layout (mirror `paperless/cli/`):
```
cli/
├── main.go                    # cobra CLI: install, configure, pre-refresh, post-refresh, storage-change, access-change
└── installer/
    ├── installer.go           # Install, Configure, Upgrade, UpdateConfigs (incl. OIDC register)
    ├── database.go            # SQLite init for peers table
    ├── executor.go
    └── oidc.go                # OpenIDConfig(...) helper — builds the OIDC config the web backend reads
```

Responsibilities:
- Create user `amneziawg`.
- Lay out `/var/snap/amneziawg/current/{config,awg}` and `/var/snap/amneziawg/common/{log,db,nginx}`.
- Generate server keypair (Curve25519 — use `awg genkey` / `awg pubkey`) on first install only.
- Generate obfuscation parameters (random `Jc 3..10`, `Jmin 50`, `Jmax 1000`, random `S1 15..150`, random `S2 15..150`, random `H1..H4` ≥ 5). Persist them — same values must end up in every client config.
- Template `awg0.conf` from `config/awg-server.conf.tpl` using generated keys + obfuscation.
- **Register as OIDC client with the platform** via `platformClient.RegisterOIDCClient(...)` — see "Auth" section below. Call this from `UpdateConfigs()` so it re-registers on every configure/refresh (paperless pattern).
- Initialize SQLite (peers, client metadata).

Where OpenVPN uses easyrsa / OpenSSL PKI, AmneziaWG just needs Curve25519 keypairs per peer — much simpler.

### Auth — OIDC via the platform (Authelia)

**Use OIDC, not forward-auth cookies.** Platform exposes a full OpenID Connect provider (Authelia). Working example: `../paperless/cli/installer/installer.go:275`.

Installer side (once per install/refresh):
```go
password, err := i.platformClient.RegisterOIDCClient(
    App,                                       // "amneziawg"
    "/auth/callback",                          // redirect path on the app's domain
    true,                                      // require PKCE
    "client_secret_basic",
)
// Persist `password` (the client_secret) where the web backend can read it.
```

Web backend side (Go):
- Discover OIDC endpoints from `${authUrl}/.well-known/openid-configuration` (where `authUrl = platformClient.GetAppUrl("auth")`).
- Standard auth-code + PKCE flow using `golang.org/x/oauth2` + `github.com/coreos/go-oidc/v3`.
- On callback, validate ID token, mint a session cookie (signed/encrypted), redirect to SPA.
- SPA calls `/api/*` with the session cookie; backend middleware checks it.

Note: paperless delegates OIDC to its upstream app (Paperless-ngx has native OIDC). amneziawg has no upstream app — we implement the OIDC client directly in our Go backend.

### Server config template

`config/awg-server.conf.tpl`:

```
[Interface]
PrivateKey = {{ .ServerPrivateKey }}
ListenPort = {{ .ListenPort }}
Address    = 10.9.0.1/24, fd42:42:42::1/64
MTU        = 1420

# AmneziaWG obfuscation
Jc = {{ .Jc }}
Jmin = {{ .Jmin }}
Jmax = {{ .Jmax }}
S1 = {{ .S1 }}
S2 = {{ .S2 }}
H1 = {{ .H1 }}
H2 = {{ .H2 }}
H3 = {{ .H3 }}
H4 = {{ .H4 }}

PostUp   = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

{{ range .Peers }}
[Peer]
# name: {{ .Name }}
PublicKey  = {{ .PublicKey }}
AllowedIPs = {{ .AllowedIPs }}
{{ end }}
```

Peers are appended/removed dynamically by the web UI; regenerate the file and `awg syncconf awg0 <(awg-quick strip awg0)` to hot-apply without dropping peers.

### Client config generated by web UI

```
[Interface]
PrivateKey = <client priv>
Address    = 10.9.0.N/32, fd42:42:42::N/128
DNS        = 10.9.0.1

Jc = ...
Jmin = ...
Jmax = ...
S1 = ...
S2 = ...
H1 = ...
H2 = ...
H3 = ...
H4 = ...

[Peer]
PublicKey  = <server pub>
Endpoint   = <device-domain>:<port>
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
```

Serve as:
- downloadable `<peer>.conf`
- QR code PNG for the Amnezia Android/iOS app
- "Open in Amnezia" deep link if feasible

### UDP port — dynamic per install, user forwards on router

**Dynamic port allocation, chosen at install time.** Other Syncloud apps may have already grabbed a given port, so hard-coding a default is fragile. On first install:
1. Pick a random high UDP port (avoid 51820 — DPI-flagged).
2. Attempt to bind to verify it's free; retry on `EADDRINUSE`.
3. Persist the chosen port in `settings` table + `awg0.conf`.
4. **Display it prominently on the Dashboard** with a "forward UDP port N on your router" instruction — this is a user concern for now; no platform-side auto-forwarding.

Single UDP port serves all peers (WG demuxes by public key, not by port).

Later enhancement: let the user edit the port in Settings (requires service restart). UDP 443 "hide in plain sight" is a nice option if free.

### Obfuscation parameter defaults

Generate fresh per install so every Syncloud device has a unique fingerprint:
- `Jc`: random 3–10
- `Jmin`: 50
- `Jmax`: 1000
- `S1`: random 15–150
- `S2`: random 15–150, with `S1 + 56 ≠ S2` (AmneziaWG requirement)
- `H1..H4`: random uint32 ≥ 5, all distinct and ≠ 1/2/3/4 (reserved WG message types)

Reference: https://github.com/amnezia-vpn/amneziawg-linux-kernel-module#parameters

### IPv6

The existing openvpn app handles IPv6 via `prefix_delegation.sh`. Reuse the same approach for AmneziaWG — DHCPv6 prefix delegation gives the device a routable /60-or-so, carve out a /64 for the VPN subnet. See `../openvpn/bin/prefix_delegation.sh`.

---

## Bundled binaries

**Zero OS library dependency — everything inside the snap must be statically linked.** Mirror `../owncast/backend/build.sh` flags: `-ldflags "-linkmode external -extldflags -static"`.

For our own Go code (installer, web backend), prefer **`CGO_ENABLED=0`** — pure static, no libc dependency at all. This forces one library swap:
- **SQLite driver: use `modernc.org/sqlite` (pure Go), NOT `mattn/go-sqlite3` (cgo).** Drop-in `database/sql` driver, zero cgo.
- `amneziawg-go`, `golib/platform`, `golang.org/x/oauth2`, `github.com/coreos/go-oidc/v3` — already pure Go, no action needed.

Needed inside the snap:
- `amneziawg-go` — userspace data plane (pure Go, `CGO_ENABLED=0`). Repo: `github.com/amnezia-vpn/amneziawg-go`.
- `awg` — CLI from `amneziawg-tools` (C — static build with `-extldflags -static`). Repo: `github.com/amnezia-vpn/amneziawg-tools`.
- `awg-quick` — wrapper script (bash).
- `iptables` — static build if not available in base.
- nginx — static build (same as owncast).
- Our own `cli` binary and `backend` binary — `CGO_ENABLED=0 go build`.

No Python runtime, no OpenSSL, no easyrsa (we drop PKI entirely).

### Target architectures

**Two archs confirmed, three if CI agrees.** All current Syncloud sibling apps (owncast, owntracks, paperless, redirect) ship `amd64 + arm64` only. `armhf` (arm/v7) has been dropped by newer apps. For amneziawg: start with `amd64 + arm64`, attempt `armhf` as a third target — if CI breaks on armhf we drop it without blocking the app.

---

## Web UI — concrete TODO

Do **not** fork `../openvpn/web` — its Go/Beego HTML-template pattern is the old style. Build two pieces:

### Backend (`backend/`, Go, API-only, listens on Unix socket)

Layout mirrors `../owncast/backend/`:
```
backend/
├── main.go                  # serves on unix:/var/snap/amneziawg/current/backend.sock
├── build.sh                 # CGO_ENABLED=0 go build ...
├── auth/
│   └── oidc.go              # OIDC client (discovery + code+PKCE + session cookie)
├── peers/
│   ├── handlers.go          # GET/POST/DELETE /api/peers, /api/peers/:id/config, /api/peers/:id/qr
│   ├── service.go           # keypair gen, next-free-IP allocation, awg syncconf
│   └── db.go                # modernc.org/sqlite
├── status/
│   └── handlers.go          # GET /api/status — parses `awg show awg0 dump`
├── settings/
│   └── handlers.go          # GET/PUT /api/settings — port, obfuscation params (PUT requires service restart)
└── logs/
    └── handlers.go          # GET /api/logs — tails journalctl -u snap.amneziawg.server
```

**No HTML templates, no server-side rendering.** Backend is a pure JSON API + static-file server for the SPA bundle.

### Frontend (`www/`, Vue 3 + Element Plus + Vite + TypeScript SPA)

Mirror `../redirect/www/` (which itself matches `../platform/web/platform/`):
```
www/
├── package.json             # vue, element-plus, vite, pinia, vue-router, typescript
├── vite.config.js           # proxy /api to the backend unix socket during dev
├── playwright.config.js     # e2e config
├── src/
│   ├── App.vue
│   ├── main.ts
│   ├── router.ts
│   ├── stores/              # pinia
│   ├── views/
│   │   ├── Dashboard.vue    # status + "forward UDP port N" callout
│   │   ├── Peers.vue        # list, add, revoke, download .conf, show QR
│   │   ├── Settings.vue     # port + obfuscation params
│   │   └── Logs.vue
│   └── components/
├── e2e/                     # Playwright specs (see Testing section)
└── tests/                   # Jest unit tests
```

### Auth — OIDC via the platform (Authelia)

**OIDC, not forward-auth cookies.** Platform exposes an OpenID Connect provider via Authelia. Working reference: `../paperless/cli/installer/installer.go:275`.

Installer side (Go, once per install/refresh):
```go
password, err := i.platformClient.RegisterOIDCClient(
    App,                                       // "amneziawg"
    "/auth/callback",                          // redirect path on the app's domain
    true,                                      // require PKCE
    "client_secret_basic",
)
```

Backend side (Go):
- OIDC discovery against `${authUrl}/.well-known/openid-configuration` (authUrl = `platformClient.GetAppUrl("auth")`).
- Standard auth-code + PKCE using `golang.org/x/oauth2` + `github.com/coreos/go-oidc/v3`.
- On callback, validate ID token → mint signed session cookie → redirect to SPA.
- Middleware verifies session cookie on `/api/*`.

Note: no other sibling app currently uses OIDC directly — paperless delegates to its upstream's native OIDC. amneziawg has no upstream to delegate to, so we implement the OIDC client in our backend.

### DB

SQLite (driver: **`modernc.org/sqlite`** — pure Go, `CGO_ENABLED=0` compatible) in `/var/snap/amneziawg/common/db/data.db`. Schema:
```
peers (id INTEGER PK, name TEXT UNIQUE, public_key TEXT, private_key TEXT, address_v4 TEXT, address_v6 TEXT, created_at TEXT, revoked INTEGER DEFAULT 0)
settings (key TEXT PK, value TEXT)
```

Don't persist private keys if you don't have to — generate client priv on the client side via the Amnezia app, only store the public key. Alternative: generate on server and deliver priv in the config once, then delete from DB. Pick one and document.

---

## Packaging — `package.sh`

Mirror `../owncast/package.sh` (newer pattern). Copy list: `cli` binary, `backend` binary, `www/dist/` SPA bundle, `amneziawg-go`, `awg`, `awg-quick`, static nginx, static iptables if needed.

---

## Testing

Three tiers, mirroring the split that newer Syncloud apps use.

### Tier 1 — Unit tests (Go)
Standard `go test ./...` across `cli/` and `backend/`. Obfuscation-param generator must satisfy constraints (`S1+56 != S2`, `H1..H4` distinct and ≥ 5).

### Tier 2 — Playwright UI tests (`www/e2e/`, mirror `../redirect/www/e2e/`)
- Config: `www/playwright.config.js` — mirror `../redirect/www/playwright.config.js` shape (desktop + mobile projects, `baseURL = https://amneziawg.${PLAYWRIGHT_DOMAIN}`, workers=1, retries=1 in CI, HTML report).
- Specs cover: OIDC login, create peer, download `.conf`, delete peer, settings edit.
- CI runs in `mcr.microsoft.com/playwright:v1.59.1-jammy` via `ci/ui.sh desktop` (reuse redirect's shell script shape).

### Tier 3 — End-to-end VPN handshake test
The UI test can verify the UI, but only a real tunnel proves the exported config actually works. So after Playwright creates a peer and downloads the `.conf`:
1. Launch a second container with `amneziawg-go` + `awg-quick` installed.
2. Place the downloaded `.conf` at `/etc/amnezia/amneziawg/client.conf`.
3. `awg-quick up client.conf`.
4. `ping 10.9.0.1` across the tunnel (expect success).
5. On the server side, `awg show awg0 dump` — assert this peer has recent handshake timestamp and non-zero `rx_bytes`/`tx_bytes`.
6. `awg-quick down client.conf`.

This is the **"can users actually VPN in after using the UI?"** acceptance test.

### Distros — test on bookworm AND buster (bitwarden pattern)

Old devices still run Debian buster, so we can't only target bookworm. Mirror `../bitwarden/.drone.jsonnet`:

```jsonnet
local distros = ["bookworm", "buster"];
local platform_buster = '25.02';  // pinned: newer platform dropped buster support
// ...
image: 'syncloud/platform-' + distro + '-' + arch + ':' +
       (if distro == 'buster' then platform_buster else platform),
```

CI matrix: `{amd64, arm64} × {bookworm, buster}` minimum. Attempt armhf as a third arch; if CI breaks we drop it.

---

## Out of scope for v1

- Management via mobile app deep-links.
- Automatic obfuscation param rotation (important long-term vs TSPU, but not v1).
- Split tunneling UI.
- Peer expiry / time-limited configs.
- Multi-server / HA.

---

## References

- AmneziaWG kernel module: https://github.com/amnezia-vpn/amneziawg-linux-kernel-module
- amneziawg-go (userspace): https://github.com/amnezia-vpn/amneziawg-go
- amneziawg-tools: https://github.com/amnezia-vpn/amneziawg-tools
- Amnezia client apps: https://github.com/amnezia-vpn/amnezia-client
- Syncloud platform docs: https://docs.syncloud.org/developers/app.html
- Sibling reference apps:
  - `../owncast` — canonical modern shape (Go CLI, Unix sockets, static build)
  - `../owntracks` — second modern-shape reference
  - `../paperless` — OIDC auth pattern (`cli/installer/installer.go:275`)
  - `../redirect` — Playwright UI test pattern (`www/e2e/`, `www/playwright.config.js`)
  - `../bitwarden` — buster + bookworm CI matrix (`.drone.jsonnet`)
  - `../openvpn` — VPN-specific bits only (IPv6 `prefix_delegation.sh`, server-config template shape). Architecture is outdated; don't mirror it.

---

## Suggested implementation order

1. **Scaffold** — copy `../owncast` skeleton (`snap.yaml`, `package.sh`, `cli/`, `bin/service.*.sh`, `config/nginx.conf`); replace owncast-specific bits with placeholders; make it build an empty snap.
2. **Bundle awg binaries** — static builds of `amneziawg-go` (pure Go, `CGO_ENABLED=0`) + `awg` (C, `-static`) + `awg-quick` (bash). Verify they run inside the snap on both amd64 and arm64.
3. **Installer (Go)** — generate keys + obfuscation params, pick + persist dynamic UDP port, write initial `awg0.conf`, call `platformClient.RegisterOIDCClient(...)`, start service, confirm `awg show awg0` works.
4. **Backend + OIDC** — Go backend on Unix socket, OIDC auth-code+PKCE flow, session cookie middleware. Manually curl through nginx to verify.
5. **Minimal Vue 3 + Element Plus frontend** — scaffold with Vite, wire to `/api`, single "add peer" flow that returns `.conf`. First end-to-end VPN handshake test here.
6. **Peer management UI** — list, revoke, QR (use `skip2/go-qrcode` for PNG generation), live status.
7. **Settings UI** — editable port + obfuscation params.
8. **IPv6** — port `../openvpn/bin/prefix_delegation.sh`.
9. **Testing** — Playwright suite + end-to-end VPN tunnel test + buster/bookworm CI matrix.
10. **Polish** — icons, branding, README.

## Open questions to resolve early

- Does `golib/platform` expose a port-allocation helper (something like `AllocateUDPPort`)? If not, the installer picks a random high port and retries on `EADDRINUSE` — acceptable but worth confirming a platform convention doesn't already exist.
- OIDC callback path — does Authelia pin a specific path shape or accept any registered redirect URI? Paperless uses `/accounts/oidc/authelia/login/callback/` (that's its upstream app's path); amneziawg can use anything, e.g. `/auth/callback`.
- armhf feasibility — will CI ship armhf for a static C build of `awg` and cgo-free `amneziawg-go`? Attempt, treat as optional.
