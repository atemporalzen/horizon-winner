# Security and responsible use

Use this repository only for authorized testing. Start with the synthetic fixture in an isolated browser profile. Do not point it at credential stores, cloud metadata, private services or other people's devices without explicit authorization.

Keep administration loopback-only and protect its bearer token. Public entry HTTP and DNS are experiment surfaces; restrict their deployment to your test environment. The DNS server is authoritative only, never recursive. Its managed targets and session capacity are configuration-controlled; it does not provide per-client rate limiting or production DNS anti-abuse controls.

Validate exact Host/authority at a local service before processing requests, reject unknown names, and require appropriate authentication. Apply those checks at every ingress, including reverse proxies, and validate Origin for browser-sensitive state changes. Loopback binding and ordinary CORS configuration alone should not be treated as sufficient protection against rebinding. COOP can affect window access but is not a general network-access defense.

A navigation may execute target scripts and send browser-managed cookies. Output and JSON reports can contain target content. They remain local by default; do not publish them without reviewing them. The byte cap applies to exported content and comparison reads, not the browser's own network/DOM memory consumption.

Report defects privately using GitHub's private vulnerability-reporting feature if enabled by the repository owner. Otherwise contact the maintainer through the contact information they publish on their GitHub profile; do not post live secrets or working reports containing sensitive target data in public issues. No service-level response guarantee is implied.
