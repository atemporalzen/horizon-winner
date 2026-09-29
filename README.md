# horizon-winner

A source-based successor to [horizon](https://github.com/atemporalzen/horizon) and [horizon-lna](https://github.com/atemporalzen/horizon-lna), centered on the **top-level popup navigation** DNS-rebinding technique discussed in [HackerNotes Ep. 193](https://blog.criticalthinkingpodcast.io/p/hackernotes-ep-193-browser-quirks-galore-with-j0r1an).

The familiar `html/amaze.html` entry point remains. This generation adds readable source, a standalone page, proof-driven outcomes, explicit target configuration, bounded runs, a synthetic fixture, IPv4/IPv6 authoritative DNS, authenticated administration, tests, and repeatable cross-platform builds. It is an original implementation, not an opaque replacement Singularity binary. See [migration](docs/migration.md) for replaced and omitted legacy features.

Use only with systems and browser profiles you are authorized to test. **An implementation of the reported technique is not a verified universal browser bypass.** The automated browser tests simulate transport; real public-to-local DNS/LNA validation is a separate operator-run step.

## Host just amaze.html

For an existing DNS/HTTP setup, use **[`html/amaze.html`](html/amaze.html)**. The root `amaze.html` is an identical convenience copy. Each is self-contained: no Go process, API, npm runtime, third-party assets, origin-trial token, or other JavaScript files are required.

1. Edit its `CONFIG` block; set `configured: true`.
2. Set `attackHostIPAddress`, `attackHostDomain`, `targetHostIPAddress`, `targetPort`, `targetPath`, and `proofMarker`. Defaults are documentation/test values, not live infrastructure.
3. If your DNS provider already allocated a hostname, set `rebindingHost` to that exact hostname. Otherwise the page constructs Singularity's hexadecimal-IP `s-…-fs-e.…` hostname with a fresh session identifier.
4. Serve this same file on the rebinding hostname **at the target's port**. Open the generated rebinding-origin link if initially visiting another host.
5. Click **Start navigation experiment** on that origin. Keep the experiment window and its visible popup open. Your external DNS authority must independently switch that hostname from the public entry IP to your controlled target IP.
6. Download the local JSON report. Correlate it with browser console/network errors, DNS answers and target logs before attributing a fetch rejection to LNA.

Use a distinctive, non-secret marker returned by the controlled target in visible document text. The default synthetic fixture returns `horizon-lab-proof-v1` at `/proof`. A rejected fetch alone never counts as success. The page appends cache-busting query parameters, so the target must accept those parameters. There is no automatic exfiltration or post-navigation command channel.

The generated hostname matches the upstream LNA branch's hex-IP convention. Compatibility with any particular Singularity deployment still depends on its resolver, strategy timing and virtual-host routing. For IPv6 in standalone mode, supply a preallocated `rebindingHost` explicitly. Do not reuse a controller-managed session hostname in standalone mode without separately arming it: standalone mode never calls the controller API.

## Optional source-based DNS/controller service

Requirements: Go 1.27.1 or newer. Node/npm are needed only for development/tests and regenerating the standalone artifact.

```sh
go mod download
make build
./bin/horizon-winner -config configs/local.example.json -check
./bin/horizon-winner -config configs/local.example.json
```

Open the printed loopback control URL, enter the one-time admin token printed at startup, select a configured target, and create a session. The admin token is held only in page memory. Set `HORIZON_ADMIN_TOKEN` to a securely generated value of at least 24 characters for a stable token; do not commit it. The `/run` session URL is the controller-managed equivalent of the standalone experiment page.

Local defaults bind HTTP to `127.0.0.1:8080`, administration to `127.0.0.1:9090`, and DNS to `127.0.0.1:5353`. They **do not** configure your browser's DNS, and the loopback entry does **not** cross a public-to-local LNA boundary. The `.test` session names need an appropriate resolver to work in a browser. Automated tests handle their own simulated transport.

For a real controlled test, adapt [`configs/public.example.json`](configs/public.example.json): replace the reserved `203.0.113.10` and example domain with your infrastructure, delegate the zone to its nameserver with appropriate glue, and make UDP/TCP 53 and the configured HTTP port reachable. Keep the admin listener on loopback; use an SSH tunnel for remote administration. Do not disable the machine's normal DNS resolver. Avoid HTTP reverse proxies that retain connections, inject COOP, rewrite Host, or upgrade to HTTPS; those can change the experiment. See [the operating guide](WIKI.md).

The target runs on the **test browser's network**, not necessarily on the public DNS/HTTP machine. In a public-entry configuration, `127.0.0.1` means the browser computer. The public entry and target must use the same scheme, hostname and URL port. HTTPS is not supported by this harness: certificate identity and target TLS requirements are a separate constraint.

## Synthetic target and defense comparison

Run on the browser machine using a matching target IP/port:

```sh
# Local defaults: entry 127.0.0.1; synthetic target 127.0.0.2
./bin/lab-target -listen 127.0.0.2:8080

# Alternative run with exact Host validation enabled
./bin/lab-target -listen 127.0.0.2:8080 -allowed-hosts '127.0.0.2:8080,localhost:8080'
```

Stop the first fixture before starting the second. The second rejects a rebound domain Host with the synthetic marker `horizon-host-denied`. The fixture serves synthetic data only. `-sever-opener` exercises COOP separation; it is a compatibility test, not a substitute for service authentication or Host validation.

## Components

| Component | Purpose |
| --- | --- |
| `html/amaze.html`, `amaze.html` | Single-file operator-hosted navigation experiment |
| `web/amaze.template.html`, `scripts/build-standalone.mjs` | Standalone artifact source and repeatable generator |
| `cmd/horizon-winner` | Optional authoritative DNS, entry HTTP and loopback admin service |
| `internal/config` | Strict validated configuration and explicit targets |
| `internal/session` | Expiring, capacity-limited, race-safe session state and idempotent arming |
| `internal/dnsserver` | UDP/TCP A/AAAA, NS/SOA, zero-TTL answers, zone and session validation |
| `internal/httpserver` | Separate admin/experiment listeners, authentication, origin checks and embedded assets |
| `web/engine.js` | Shared bounded state machine and evidence classification |
| `web/runner.js`, `web/controller.js` | Managed experiment and controller UI |
| `cmd/lab-target` | Synthetic vulnerable/protected service fixture |
| `tests`, `.github/workflows`, `scripts/release.sh` | Regression tests, CI, checksums, SBOM and provenance release workflow |

Managed DNS changes from entry to target after the operator starts and the server receives the authenticated arm request. A/AAAA probes and duplicate arm requests do not alter the schedule. Session identifiers have 96 bits of randomness, run capabilities 192 bits; full hostnames are bound to the server-selected target. The server is not a recursive resolver or arbitrary-target hostname decoder.

## Results and limitations

| Reported outcome | Evidence, and what it does not prove |
| --- | --- |
| `navigation_after_fetch_rejection` | Same-origin popup document contained the configured marker after a fetch rejection; JavaScript cannot establish the rejection's cause |
| `navigation_confirmed` | Marker read through the popup; no fetch-blocking bypass established |
| `host_header_protected` | Controlled fixture's explicit Host-denial marker was read; not a detector for all defenses |
| `popup_closed_or_severed` | Popup reference unavailable; close and COOP separation may be indistinguishable |
| `timed_out` | No marker confirmed before the deadline; not proof the target is safe |
| `stopped`, `navigation_error`, `setup_error` | Cancelled run or explicit setup/navigation failure |

DNS minimum TTLs, OS/browser/DoH caches, connection reuse, proxies, permission decisions, popup policies, redirects, authentication and target response headers can all affect results. Target navigation loads an actual document and may execute its scripts or send browser-managed cookies: use a dedicated test profile and a synthetic target. Reports may contain target content; keep them private. Fetch comparisons omit credentials, but top-level navigation cannot be forced to do so.

`maxResponseBytes` bounds comparison reads and exported UTF-8 document content. It cannot cap the browser's own loading, DOM construction, or full `outerHTML` serialization. This is a document-read experiment, not an unrestricted local-network proxy.

## Development and releases

```sh
npm ci --ignore-scripts
npm run build:standalone
make test check
npx playwright install chrome firefox
npm run test:browser
./scripts/release.sh v0.1.0
```

Regenerating overwrites both standalone copies: edit the template for repository development, or edit the generated file only for deployment configuration. Tests check that the committed artifacts match their sources. Browser tests use real browser engines with **simulated** entry/target responses and blocked fetches; they are explicitly not real DNS or LNA policy tests.

Release builds cover Linux, macOS and Windows on amd64/arm64. Source, lockfiles, the exact Go toolchain, version and build flags are needed to reproduce hashes. CI actions are commit-pinned and Dependabot tracks updates. Publishing a `v*` tag runs the release workflow; local build checks do not establish that hosted GitHub Actions passed.

Read [Namecheap/DigitalOcean setup](docs/setup.md), [research and fork findings](docs/research.md), [migration](docs/migration.md), [validation status](docs/verification.md), and [security notes](SECURITY.md). Licensed under MIT; see [NOTICE](NOTICE.md) for provenance.
