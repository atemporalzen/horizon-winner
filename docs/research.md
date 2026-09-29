# Research and upstream survey

Snapshot date: 2026-09-29. Claims below separate source documentation from independent validation.

## Navigation research

[HackerNotes Ep. 193, published September 24, 2026](https://blog.criticalthinkingpodcast.io/p/hackernotes-ep-193-browser-quirks-galore-with-j0r1an) reports Jorian Woltjer's DNS-rebinding popup-navigation technique. Horizon Winner implements the core mechanism: open a same-origin popup in a user gesture, wait for DNS transition or observe a rejected fetch, navigate the popup, and read its document through the retained window reference.

The reported rationale is supported for Chrome by [Chrome 147's official release notes](https://developer.chrome.com/release-notes/147), which explicitly state that main-frame navigation is not currently subject to LNA restrictions. This statement is not a guarantee for future versions or every target. Chrome's [LNA overview](https://developer.chrome.com/blog/local-network-access) explains the permission model; subsequent versions add and change enforcement surfaces.

The excerpt's claim about Firefox shipping equivalent LNA behavior, and its assertion that no relevant browser work occurred since January, were not independently established from Mozilla/Chromium issue records in this review. They must not be represented as verified compatibility or current bug-tracker status. Real Firefox public-to-local testing remains required.

No origin-trial token is used here. Upstream's [current README](https://github.com/nccgroup/singularity/blob/master/Readme.md) describes the non-secure-context trial as ending May 18, 2026. A permission/trial exception is distinct from the navigation enforcement boundary; Horizon-lna's old setup claims must not be carried over as timeless guidance.

## Previous generations

| Source | Inspected snapshot | Findings |
| --- | --- | --- |
| [horizon](https://github.com/atemporalzen/horizon/tree/fcfbdeb204677040cb974b6c4d0bbc34851e1b43) | `fcfbdeb204677040cb974b6c4d0bbc34851e1b43` | Small HTML/payload distribution plus Linux amd64 executable; no corresponding Go source/build manifest; inspected binary identified Go 1.20.6 and upstream revision `4fdbbb88e2c84a079b634a2830a3aaf1a751b276` |
| [horizon-lna](https://github.com/atemporalzen/horizon-lna/tree/2112f486b808d3089d1912253ec287b23938d114) | `2112f486b808d3089d1912253ec287b23938d114` | Hex-IP hostname update and newer executable; inspected binary identified Go 1.26.1 and upstream experimental revision `4301f9da6a69a212adbf72669d8d63f1477d6340`; README relied on origin-trial configuration |
| [Singularity master](https://github.com/nccgroup/singularity/tree/4b9aa4fba90ee5f3539f43e8b8f4a2f8503d97d2) | `4b9aa4fba90ee5f3539f43e8b8f4a2f8503d97d2` | Newer dependency, concurrency, missing-session and IPv6 fixes absent from the old packaged snapshot |

The predecessor review also found wildcard message handling, simplistic success detection, copied setup guidance, loading payload files unrelated to selection, external collection defaults, and a lack of automated tests/CI/license metadata. The migration matrix records the response to each; this rewrite does not assert that all possible predecessor defects have been found.

## Singularity fork/PR findings

The review sampled the 100 newest forks, examined recently active candidates, compared default-branch ancestry against upstream, and inspected relevant open pull requests. This is not an exhaustive audit of every branch in every fork. A fork's recent push date or repository size does not establish new default-branch features.

| Candidate | Default-branch comparison at review time |
| --- | --- |
| [boosters-research](https://github.com/boosters-research/singularity-dns-rebinding), [PwnDexter](https://github.com/PwnDexter/singularity) | No default-branch divergence from inspected upstream |
| [falasi](https://github.com/falasi/singularity), [zamibd](https://github.com/zamibd/singularity), [nvroot](https://github.com/nvroot/singularity) | Behind upstream by two commits; no ahead commits in the comparison |
| [ozkalkans](https://github.com/ozkalkans/singularity) | Behind by six |
| [ttttmr](https://github.com/ttttmr/singularity) | Behind by 52; recency alone was not evidence of default-branch additions |
| [JLLeitschuh](https://github.com/JLLeitschuh/singularity) | Default branch behind by 18; useful work existed on separate PR branches |

[PR #74](https://github.com/nccgroup/singularity/pull/74), a draft at review time, introduces configurable headers and JSON-schema payload configuration. These are useful extension ideas, not merged guarantees, and were not imported. Browser API-controlled headers cannot set forbidden headers such as `Host`; the bypass depends on the original domain Host remaining unchanged.

[PR #47](https://github.com/nccgroup/singularity/pull/47) discusses macOS platform/privilege handling and alternate IP encodings. This rewrite avoids the platform-specific old server implementation and has static Go cross-build targets. [PR #71](https://github.com/nccgroup/singularity/pull/71) and [PR #75](https://github.com/nccgroup/singularity/pull/75) concern target-specific MCP exploitation; those payloads are outside the synthetic navigation validation implementation and were not imported.

The new service uses locked state, expiry-aware lookups and standard `net` listeners rather than claiming upstream patches were cherry-picked into a vendored binary. IPv6 session labels are explicitly tested against DNS's 63-byte label limit. Future imports must pin exact revisions, preserve licensing and add tests before being labeled supported.
