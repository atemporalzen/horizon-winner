# Provenance

The Horizon-compatible distribution restores the HTML/payload workflow from atemporalzen/horizon-lna at 2112f486b808d3089d1912253ec287b23938d114, with navigation, configuration, message-validation and payload-loading changes.

server/ vendors NCC Group Singularity at 4b9aa4fba90ee5f3539f43e8b8f4a2f8503d97d2. Its MIT notice is preserved in server/LICENSE. The Go-derived domain-validation helper's license is preserved in server/golang/LICENSE.

Local server changes: selected-payload loading rather than concatenating every payload; marked entry pages; -IPAddress alias; optional -DNSServerPort; platform-specific transparent-proxy helpers; updated Go/dependencies. Attack settings remain in html/amaze.html.

The original controller-based implementation introduced in the first horizon-winner commit remains available but is not required for the default Horizon-compatible workflow. Its MIT license applies to the original code, not as a replacement for upstream notices.

Navigation research: Jorian Woltjer, as discussed in HackerNotes Ep. 193 on September 24, 2026. No expired origin-trial token is required by the new navigation method. A working browser/DNS boundary result is not inferred from a simulated test.
