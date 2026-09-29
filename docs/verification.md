# Verification record

## Horizon-compatible workflow correction

The default workflow now uses `html/amaze.html` plus the `singularity-server` launcher, not the original controller. All operator attack settings, including the collector, are in `amaze.html`. The launcher compiles the pinned vendored source on first run with Go 1.27.1; no separate build command is required.

Local checks of this correction: 21 Chrome browser tests passed, including six new Horizon-layout cases (navigation/AWS config, legacy fetch, selected Hook loading, entry-error rejection, popup blocking and cancellation). The tests use the actual local Singularity HTTP template/assets with simulated rebinding transport. The vendored server passed its Go race tests, vet and govulncheck; macOS arm64 and Linux amd64 builds succeeded. Hosted CI for this correction must be checked separately; earlier CI results below apply to the earlier source snapshot.

Real public-to-local DNS/LNA validation remains pending operator infrastructure. AWS collector transmission and the full live WebSocket command channel were not exercised against an external target in these tests.

## Earlier controller/standalone verification snapshot

Local verification date: 2026-09-29. Host: macOS arm64, Go 1.27.1, Node 20.19.1, Google Chrome 154.0.8037.58, Playwright 1.63.0.

| Check | Result |
| --- | --- |
| Go tests with race detector | Passed: configuration, sessions, HTTP authentication/origin/listener separation, authoritative DNS and live UDP/TCP exchanges |
| Go vet | Passed |
| Shared state-machine/UTF-8 unit tests | 10 passed |
| Managed and standalone Chrome browser tests | 15 passed; actual popup/window/document operations with simulated transport |
| Embedded/standalone JavaScript parse and generated-artifact consistency | Passed |
| Local and public example configuration validation | Passed |
| `govulncheck` v1.8.0 | No vulnerabilities found for the application at check time |
| npm audit | Zero reported dependency vulnerabilities at check time |
| Linux/macOS/Windows amd64/arm64 compilation | All six main-server targets built |
| Same-toolchain Linux amd64 rebuild comparison | Identical SHA-256: `922cd70b92a96a12e35eac3d17c9b3928a73ca706f1ad29f081c2f4bba195ea8` for version `v0.1.0` at the verification snapshot |
| CycloneDX Go-module SBOM | Generated locally with cyclonedx-gomod v1.12.0; main-module version warning before Git initialization |
| Firefox browser suite on this macOS host | Not executed successfully: bundled Firefox 155.0 exited during launch with “Could not find profile folder”; rerun with an operational Firefox test runtime |
| Real public-entry → loopback/private DNS/LNA boundary | **Not tested**; requires operator infrastructure and target |
| Hosted GitHub Actions | [CI run 36589374220](https://github.com/atemporalzen/horizon-winner/actions/runs/36589374220) passed both jobs for source commit `6a5431a81feb63c582eaaf7114951d299ec7a5fc`: Go/race/vet/config/generated-artifact checks, dependency vulnerability check, and Chrome/Firefox browser suites on Linux |
| Hosted release provenance | Release workflow supplied; no release tag or hosted release has been published |

The browser tests deliberately abort comparison fetches and fulfill synthetic target navigations. They prove UI/evidence behavior, not that the browser itself rejected a real fetch for LNA. Both Chrome and Firefox suites passed in Linux CI; the local macOS Firefox runtime launch failure above is a separate environment limitation. Real DNS/LNA compatibility remains unverified in either browser. Do not describe a timed-out run as a confirmed defense or a fetch exception as a confirmed LNA event.
