# Verification record

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
| Hosted GitHub Actions/release provenance | Workflow supplied; status must be checked on GitHub, not inferred from local tests |

The Chrome tests deliberately abort comparison fetches and fulfill synthetic target navigations. They prove UI/evidence behavior, not that Chrome itself rejected a real fetch for LNA. Firefox is configured in CI but compatibility is unverified until those runs complete. Do not describe a timed-out run as a confirmed defense or a fetch exception as a confirmed LNA event.
