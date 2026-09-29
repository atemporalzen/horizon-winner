# Operating guide

For the current Horizon-compatible edit-one-file workflow, follow README.md. The sections below document the earlier optional controller/standalone harness and are not required to launch `singularity-server` with `html/amaze.html`.

## Existing Horizon-style hosting

Host `html/amaze.html` as `/amaze.html` after editing its `CONFIG`. It contains the state machine, styles and browser adapter; no `payload.js` request is required. Use the exact preallocated rebinding hostname or let it construct the upstream hexadecimal IPv4 session hostname. The generated link keeps the current page path, so the same file must be available at that path on the entry server for every session hostname.

This file cannot change DNS records. The operator controls the DNS authority and switching strategy. TTL 0 is a request to caches, not a command that forces immediate re-resolution. The default 60-second wait and 180-second deadline are tunable observations, not browser timing guarantees.

The optional Go service is a different session implementation: it rejects unknown session names. Create sessions through its loopback controller and use `/run`; do not feed it a standalone-generated hostname.

## Delegated-zone checklist

1. Allocate a domain/subdomain and an entry machine you control.
2. Delegate the rebinding zone using NS records and resolvable nameserver addresses. If a nameserver is inside the delegated zone, arrange parent-zone glue as appropriate. The Go service supplies authoritative zone NS/SOA answers, not registrar configuration.
3. Make both UDP and TCP DNS reachable. Separate authoritative DNS from the machine's ordinary recursive/system resolver.
4. Serve entry HTTP on the target URL's port with routing that accepts the generated session hostnames. A path-changing redirect or domain-changing redirect can invalidate origin equality.
5. Run the synthetic fixture on the test browser computer. Use a browser profile without real account sessions or valuable extensions.
6. Verify authoritative answers before and after the operator-controlled switch, then verify the resolver actually used by the browser. DoH can bypass the OS resolver you inspected.

For the managed service, `-verbose` logs DNS session/type/answer observations, and the arm response exposes `rebind_at`. Session state is in memory; restart invalidates all sessions. Expired or forged names return NXDOMAIN, other zones REFUSED, and the opposite address family NODATA. A/AAAA queries do not advance a query-count rebinding phase.

## Browser evidence checklist

Record date, full browser version, OS, permission state for both local-network and loopback access where available, managed enterprise policies, resolver/DoH configuration, entry and target address family, target port/path, target Host behavior and response headers. Do not disable LNA/security flags to make a result look successful.

Confirm all three observations independently:

- DNS/target logs show that the rebinding hostname actually reached the configured target.
- The comparison fetch was rejected with browser evidence identifying LNA rather than timeouts, connection errors, redirects, extensions, or a JavaScript exception.
- The popup document was readable through the same-origin window reference and contained the synthetic marker.

Only that combined record supports a specific LNA-navigation bypass result. The UI's `navigation_after_fetch_rejection` outcome deliberately makes a narrower claim.

## Troubleshooting

| Symptom | Checks |
| --- | --- |
| Generated origin does not load | DNS delegation, browser resolver, entry firewall, target-port entry listener and wildcard/session Host routing |
| Managed session returns 421/NXDOMAIN | Exact host/port, unexpired session, process restart, standalone-vs-managed hostname mismatch |
| Popup blocked | Start must be a direct click; explicitly allow the visible test popup |
| Entry content persists | Browser/DoH caches, server keepalive, connection-pooling reverse proxy, DNS switch state |
| Fetch rejects before DNS changes | Insecure-context policy, extension/proxy blocks, timeout, or ordinary connection failure; not automatic LNA proof |
| Target reached but no proof | Marker not in visible text, wrong path, query parameters rejected, redirect, authentication, unreadable/severed popup |
| Private entry confirmed | This is plumbing validation, not a public-to-local boundary test |

No offscreen popup, unbounded polling, resolver disabling, external collector, recursive DNS, or hardcoded credential/metadata payload is needed for this test.
