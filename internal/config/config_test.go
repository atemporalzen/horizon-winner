package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidation(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Config)
	}{
		{"bad octet", func(c *Config) { c.Targets[0].IP = "127.0.0.999" }},
		{"mixed family", func(c *Config) { c.Targets[0].IP = "::1" }},
		{"port changes origin", func(c *Config) { c.Targets[0].Port = 9000 }},
		{"external path", func(c *Config) { c.Targets[0].Path = "//elsewhere/proof" }},
		{"backslash path", func(c *Config) { c.Targets[0].Path = "/\\elsewhere/proof" }},
		{"domain port", func(c *Config) { c.Domain = "example.com:80" }},
		{"invalid contact", func(c *Config) { c.DNSContact = "admin@example.com" }},
		{"short proof", func(c *Config) { c.Targets[0].ProofMarker = "" }},
		{"short ttl", func(c *Config) { c.SessionTTLSeconds = 1 }},
		{"unknown dns ns", func(c *Config) { c.DNSNameserver = "ns.other.test" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := Defaults()
			test.change(&c)
			if err := c.Validate(); err == nil {
				t.Fatal("invalid config accepted")
			}
		})
	}
	c := Defaults()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	c.EntryIP = "2001:db8::1"
	c.Targets[0].IP = "::1"
	c.HTTPListen = "[::1]:8080"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestLoadRejectsUnknownAndTrailingJSON(t *testing.T) {
	for _, body := range []string{`{"typo":true}`, `{} {}`, `{} garbage`} {
		p := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(p); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
	data, _ := json.Marshal(Defaults())
	p := filepath.Join(t.TempDir(), "config.json")
	_ = os.WriteFile(p, data, 0600)
	if _, err := Load(p); err != nil {
		t.Fatal(err)
	}
}
