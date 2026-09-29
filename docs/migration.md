# Migration from Horizon / Horizon-lna

| Earlier behavior or shortcoming | Horizon Winner treatment |
| --- | --- |
| `html/amaze.html` entry point | Retained as a self-contained generated artifact; root convenience copy also supplied |
| Binary-only Linux server distribution | Original Go source, locked dependencies and six-platform static build script |
| Fetch/iframe-centered flow | Visible user-gesture popup, repeated bounded top-level navigation and same-origin document read; fetch retained as comparison |
| Non-secure-context origin-trial setup | No trial dependency; source-dated research and explicit browser validation limits |
| Dotted-vs-hex IP hostname ambiguity | Standalone IPv4 builder uses upstream hex convention; preallocated exact hostname supported |
| Unbounded loops, fragile fixed delays | Deadline, request timeout, polling cadence, restrained navigation retries, Stop and popup lifecycle outcomes |
| Errors treated as rebinding success | Confirmation requires a configured marker in a readable target document; fetch cause kept unproven |
| Hidden/offscreen UI | Visible popup and operator-facing status |
| Hardcoded external collection | Local bounded document display and downloadable JSON; no external collector |
| Unrelated payload script parsing | No legacy payload auto-loading; shared tested engine and dedicated adapters |
| Wildcard postMessage trust | No cross-window message protocol; direct same-origin window access with origin/path/nonce checks |
| Arbitrary host-encoded targets | Optional server binds complete session hosts to an explicit configured target list |
| Shared or exposed administration | Separate loopback admin listener, exact Host, bearer authentication and browser origin checks |
| Session/concurrency fragility | Locked bounded state, expiry, cryptographic identifiers/capabilities and idempotent arm |
| IPv6 parsing/label issues | Typed IP handling, same-family validation, A/AAAA correctness and legal label-length tests |
| Working-directory-sensitive assets | Managed web assets embedded in Go executable; standalone page bundles its own assets |
| Resolver-disable setup steps | Removed; authoritative DNS is separately configured without replacing the system resolver |
| No reproducible build/CI trail | Trimmed static builds, checksums, commit-pinned CI, tests, dependency checks and tag-release SBOM/provenance workflow |
| No clear license/provenance | MIT for the original implementation, NOTICE and dated upstream/research references |

This is a successor, not a drop-in CLI-compatible Singularity fork. The old `singularity-server` flags, index manager, scanner, `/delaydomload`, DNS cache-flush helpers, `rr`/`rd`/`ma` strategies, persistent WebSocket hook/proxy and AWS-metadata exfiltration payload are not implemented here. The managed DNS strategy is operator-armed first-then-second, not query-count-based upstream `fs`. Standalone mode can use an existing upstream DNS authority independently.

Other fork additions—custom header/schema UI, alternate decimal encodings, MCP-specific payloads—remain explicit future work, not advertised features. There is no post-navigation `w.fetch()` proxy, authentication bypass, cookie control for navigations, TLS support, or universal-browser success claim. This scope makes the new navigation behavior independently inspectable and testable while retaining the familiar single-file hosting workflow.
