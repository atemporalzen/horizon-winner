# horizon-winner

Horizon, with the existing edit-one-file workflow preserved and the newer navigation method added.

## Your usual setup

Keep your Namecheap/DigitalOcean setup and `dynamic.your.domain` delegation. Install Go 1.27.1 or newer, then:

```sh
git clone https://github.com/atemporalzen/horizon-winner
cd horizon-winner/html
nano amaze.html
cd ..
./singularity-server -IPAddress YOUR_DROPLET_IP -HTTPServerPort 80 -HTTPServerPort 8080
```

The server needs permission to bind DNS port 53 and HTTP port 80. If you only serve HTTP on 8080, omit the port-80 flag and include `:8080` in the entry URL.

The target visits **`http://az2.website/amaze.html`** if your entry domain points to the droplet and port 80 is listening. `attackHostDomain` can remain `dynamic.az2.website`; the entry hostname and rebinding subdomain are different roles.

## All attack settings stay in html/amaze.html

```js
const CONFIG = {
    attackHostIPAddress: 'YOUR_DROPLET_IP',
    attackHostDomain:    'dynamic.az2.website',
    targetHostIPAddress: '127.0.0.1',
    targetPort:          '8080',
    targetPath:          '/',
    rebindingStrategy:   'fs',
    interval:            '20',
    attackMethod:        'navigation',
    attackPayload:       'AWS Metadata Exfil',
    exfiltrationURL:     '', // your collector URL, or blank for console-only output
    // Other limits and optional settings are in that same CONFIG block.
};
```

The entry server must also listen on `targetPort`, since the rebinding hostname keeps the same port when it switches IPs. `interval: '20'` is the polling/navigation interval in seconds, **not** the full run timeout. The latter is `runTimeoutSeconds`.

No JSON configuration file, controller login, generated controller session, or mandatory target proof marker is needed.

## Methods and payloads

- `navigation`: the new popup-navigation technique. The entry page loads the rebinding frame automatically; click **Start navigation test** inside it once to satisfy the browser popup policy. The frame retains the same-origin popup reference and reads the target document after navigation.
- `fetch`, `iframe`: the earlier methods remain selectable.
- `AWS Metadata Exfil`: retains its familiar name; sends the captured body only to the collector you explicitly set in `exfiltrationURL`. A blank URL leaves it in the console. Selecting it does not make a localhost service an AWS metadata endpoint.
- `Hook and Control`: retains the WebSocket control payload.
- `Read Response`: console-only response capture.

Only the selected payload script is loaded. Parent/frame messages check both source and exact origin. Runs have timeouts and response limits. A failed fetch is never sufficient to confirm navigation; the popup must expose a non-entry document. `proofMarker` is an optional stronger check, not a required setup step.

## Server and provenance

`singularity-server` is an executable launcher: it builds the vendored, pinned Singularity source on first run and starts it from the repository root, preserving the original static-HTML paths and flags. Go is required on the droplet; no separate manual build command is needed. The real compiled executable is cached in `.cache/`. `-IPAddress` aliases upstream's `-ResponseIPAddr`; `-HTTPServerPort` remains repeatable.

`server/` contains NCC Group Singularity at `4b9aa4fba90ee5f3539f43e8b8f4a2f8503d97d2`, including upstream session/race/IPv6 fixes, plus documented local compatibility patches. Its MIT license and the Go-derived helper's license are retained. The previously created controller implementation remains in source/history as an alternative; it is **not part of the workflow above**.

The [navigation research](https://blog.criticalthinkingpodcast.io/p/hackernotes-ep-193-browser-quirks-galore-with-j0r1an) motivated this method. Browser tests simulate entry/target responses; real public-to-local DNS/LNA validation still depends on your infrastructure, browser version and target defenses. Host checks, authentication, redirects, COOP, DNS caches and connection reuse can prevent access. The method does not guarantee access to every local service.

Use only with authorized test systems. Navigation can execute target scripts and send browser-managed cookies; use an isolated test profile. No Namecheap, DigitalOcean or system resolver settings are changed automatically.
