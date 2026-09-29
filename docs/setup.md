# Your dynamic-domain hosting workflow

The public examples retain your `dynamic.xxx.com` convention. Set the full domain you control, not the literal example. No Namecheap, DigitalOcean, resolver or system-service settings are changed by this repository.

## Option A: existing Singularity infrastructure

Clone `horizon-winner`, edit `html/amaze.html`, and host that page using your existing DNS/HTTP service. Supply `rebindingHost` if you already allocated a compatible session name. Otherwise the automatic builder uses the upstream hexadecimal-IP syntax, not the older dotted-decimal variant. Keep the `fs` authority's switching behavior under your control.

Your existing [atemporalzen/singularity](https://github.com/atemporalzen/singularity) distribution contains packaged HTML and an executable rather than this new source-based server. Installing a newer Go version does not rebuild or update an already packaged executable. Do not assume that binary accepts every hostname format or has current upstream fixes; verify its actual authoritative answers before testing the browser. Horizon Winner does not alter that separate repository or silently substitute its binary.

## Option B: the new source-based service

```sh
git clone https://github.com/atemporalzen/horizon-winner
cd horizon-winner
go version
go mod download
make build
./bin/horizon-winner -config configs/public.example.json -check
```

Go 1.27.1 was used for local verification. Install Go from its [official installation instructions](https://go.dev/doc/install), including archive checksum verification, rather than piping an unreviewed remote installer to a shell. Edit `configs/public.example.json` with your entry IP, `dynamic.xxx.com` zone and explicit synthetic target. Then run the server and use its loopback controller to create `/run` sessions. These sessions must be armed through that controller flow.

## DNS and machine resolver

`systemd-resolved` often listens on a stub loopback address; binding the authoritative server to the droplet's specific interface/IP may avoid a port conflict without disabling the resolver. Check the actual listeners and OS network configuration first. Wildcard `0.0.0.0:53` may conflict with a stub listener; the example can be adapted to a specific local interface address. A cloud public IP may be routed/NATed rather than locally bindable.

Do not blindly replace `/etc/resolv.conf` with `nameserver 8.8.8.8`: it may be a managed symlink, and replacement can break private/cloud DNS or be overwritten. Preserve the host's normal resolver through its supported network-manager settings. Recursive host resolution and authoritative test DNS are separate functions. A browser may use DoH instead of the host resolver.

## Namecheap and DigitalOcean roles

Keep registrar delegation distinct from authoritative zone records. If DigitalOcean hosts the parent zone, registrar nameservers point to that provider; the `dynamic` child-zone NS records in the parent delegate the child to your rebinding authority. That authority must have a resolvable nameserver name and reachable UDP/TCP 53. An in-child-zone nameserver needs the appropriate parent glue/address arrangements. A Namecheap personal-nameserver registration alone does not create a complete child-zone delegation.

Namecheap documents [personal nameserver registration](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/); DigitalOcean documents [DNS record management](https://docs.digitalocean.com/products/networking/dns/). Follow the provider currently authoritative for your parent zone rather than editing records at two unrelated DNS authorities. Do not alter your domain's existing mail/web records or delegate the entire domain just to test a child zone.

Keep the entry HTTP port equal to the target port. Configure routing to serve the same `amaze.html` path for every generated hostname. Use a disposable test browser profile and fixture on the browser machine. Neither ordinary DNS propagation nor a zero TTL guarantees the browser immediately changes connections.
