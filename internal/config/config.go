package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Target struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
	Path        string `json:"path"`
	ProofMarker string `json:"proof_marker"`
}

type Config struct {
	Domain             string   `json:"domain"`
	EntryIP            string   `json:"entry_ip"`
	HTTPListen         string   `json:"http_listen"`
	AdminListen        string   `json:"admin_listen"`
	DNSListen          string   `json:"dns_listen"`
	DNSNameserver      string   `json:"dns_nameserver"`
	DNSContact         string   `json:"dns_contact"`
	RebindAfterSeconds int      `json:"rebind_after_seconds"`
	SessionTTLSeconds  int      `json:"session_ttl_seconds"`
	MaxSessions        int      `json:"max_sessions"`
	PollIntervalMS     int      `json:"poll_interval_ms"`
	RequestTimeoutMS   int      `json:"request_timeout_ms"`
	RunTimeoutSeconds  int      `json:"run_timeout_seconds"`
	MaxResponseBytes   int      `json:"max_response_bytes"`
	Targets            []Target `json:"targets"`
}

func Defaults() Config {
	return Config{Domain: "rebind.test", EntryIP: "127.0.0.1", HTTPListen: "127.0.0.1:8080", AdminListen: "127.0.0.1:9090", DNSListen: "127.0.0.1:5353",
		DNSNameserver: "ns.rebind.test", DNSContact: "hostmaster.rebind.test", RebindAfterSeconds: 10, SessionTTLSeconds: 600,
		MaxSessions: 100, PollIntervalMS: 2000, RequestTimeoutMS: 4000, RunTimeoutSeconds: 180, MaxResponseBytes: 65536,
		Targets: []Target{{ID: "lab", Label: "Mock loopback service", IP: "127.0.0.2", Port: 8080, Path: "/proof", ProofMarker: "horizon-lab-proof-v1"}}}
}

func Load(path string) (Config, error) {
	c := Defaults()
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			return c, err
		}
		defer f.Close()
		d := json.NewDecoder(io.LimitReader(f, 1<<20))
		d.DisallowUnknownFields()
		if err := d.Decode(&c); err != nil {
			return c, fmt.Errorf("config: %w", err)
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return c, errors.New("config: expected one JSON object")
		}
	}
	return c, c.Validate()
}

func validDomain(d string) bool {
	if len(d) > 150 || len(d) < 3 || !strings.Contains(d, ".") {
		return false
	}
	for _, label := range strings.Split(d, ".") {
		if len(label) < 1 || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, ch := range label {
			if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-') {
				return false
			}
		}
	}
	return true
}

func (c Config) Validate() error {
	if !validDomain(c.Domain) || !validDomain(c.DNSNameserver) || !validDomain(c.DNSContact) {
		return errors.New("domain, DNS nameserver and DNS contact must be lowercase DNS names")
	}
	if !strings.HasSuffix(c.DNSNameserver, "."+c.Domain) {
		return errors.New("DNS nameserver must be inside the configured zone")
	}
	entry, err := netip.ParseAddr(c.EntryIP)
	if err != nil || entry.IsUnspecified() || entry.IsMulticast() || entry.Zone() != "" {
		return errors.New("entry_ip must be a routable IPv4/IPv6 literal without a zone")
	}
	for _, a := range []string{c.HTTPListen, c.AdminListen, c.DNSListen} {
		host, port, err := net.SplitHostPort(a)
		if err != nil {
			return fmt.Errorf("invalid listen address %q", a)
		}
		if host != "" {
			if _, err := netip.ParseAddr(host); err != nil {
				return fmt.Errorf("listen address must contain an IP literal: %q", a)
			}
		}
		p, err := strconv.Atoi(port)
		if err != nil || p < 1 || p > 65535 {
			return fmt.Errorf("invalid listen port: %q", a)
		}
	}
	adminIP, _, _ := net.SplitHostPort(c.AdminListen)
	admin, err := netip.ParseAddr(adminIP)
	if err != nil || !admin.IsLoopback() {
		return errors.New("admin_listen must bind a loopback address; use an SSH tunnel for remote administration")
	}
	if c.AdminListen == c.HTTPListen || c.AdminListen == c.DNSListen {
		return errors.New("admin listener must be distinct from HTTP and DNS listeners")
	}
	if c.RebindAfterSeconds < 1 || c.RebindAfterSeconds > 120 {
		return errors.New("rebind_after_seconds must be 1..120")
	}
	if c.RunTimeoutSeconds < 5 || c.RunTimeoutSeconds > 600 || c.RunTimeoutSeconds <= c.RebindAfterSeconds {
		return errors.New("run timeout must be 5..600 seconds and longer than DNS hold")
	}
	if c.SessionTTLSeconds < c.RunTimeoutSeconds+c.RebindAfterSeconds || c.SessionTTLSeconds > 3600 {
		return errors.New("session TTL must cover the run and DNS hold, up to 3600 seconds")
	}
	if c.MaxSessions < 1 || c.MaxSessions > 1000 {
		return errors.New("max_sessions must be 1..1000")
	}
	if c.PollIntervalMS < 250 || c.PollIntervalMS > 30000 || c.RequestTimeoutMS < 250 || c.RequestTimeoutMS > 30000 {
		return errors.New("poll interval and request timeout must be 250..30000 ms")
	}
	if c.MaxResponseBytes < 256 || c.MaxResponseBytes > 1048576 {
		return errors.New("max_response_bytes must be 256..1048576")
	}
	if len(c.Targets) == 0 || len(c.Targets) > 50 {
		return errors.New("configure 1..50 explicit targets")
	}
	seen := map[string]bool{}
	for _, t := range c.Targets {
		if len(t.ID) < 1 || len(t.ID) > 32 || seen[t.ID] {
			return errors.New("target IDs must be unique, 1..32 characters")
		}
		for _, ch := range t.ID {
			if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-') {
				return errors.New("target IDs must contain lowercase letters, digits or hyphens")
			}
		}
		seen[t.ID] = true
		if len(t.Label) < 1 || len(t.Label) > 100 {
			return errors.New("target label must be 1..100 characters")
		}
		ip, err := netip.ParseAddr(t.IP)
		if err != nil || ip.IsUnspecified() || ip.IsMulticast() || ip.Zone() != "" {
			return fmt.Errorf("target %s requires an IPv4/IPv6 literal without a zone", t.ID)
		}
		if ip.Unmap().Is4() != entry.Unmap().Is4() {
			return errors.New("entry and target must use the same address family")
		}
		if t.Port != c.HTTPPort() {
			return fmt.Errorf("target %s port must equal HTTP listener port %d: browser origins include the port", t.ID, c.HTTPPort())
		}
		u, err := url.Parse(t.Path)
		if err != nil || !strings.HasPrefix(t.Path, "/") || strings.HasPrefix(t.Path, "//") || u.IsAbs() || u.Host != "" || u.Fragment != "" || u.User != nil || strings.ContainsAny(t.Path, "\r\n\\") {
			return fmt.Errorf("target %s requires an absolute same-origin path", t.ID)
		}
		if len(t.Path) > 2048 || len(t.ProofMarker) < 8 || len(t.ProofMarker) > 128 {
			return errors.New("target path max 2048 chars; proof marker must be 8..128 chars")
		}
	}
	return nil
}

func (c Config) HTTPPort() int {
	_, p, _ := net.SplitHostPort(c.HTTPListen)
	n, _ := strconv.Atoi(p)
	return n
}
func (c Config) Host(name string) string {
	if c.HTTPPort() == 80 {
		return name
	}
	return net.JoinHostPort(name, strconv.Itoa(c.HTTPPort()))
}
func (c Config) ControlHost() string { return c.AdminListen }
func (c Config) ControlURL() string  { return "http://" + c.ControlHost() }
func (c Config) Target(id string) (Target, bool) {
	for _, t := range c.Targets {
		if t.ID == id {
			return t, true
		}
	}
	return Target{}, false
}
