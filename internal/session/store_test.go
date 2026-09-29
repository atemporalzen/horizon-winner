package session

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/atemporalzen/horizon-winner/internal/config"
)

func TestArmExpiryAndHostBinding(t *testing.T) {
	c := config.Defaults()
	c.MaxSessions = 1
	s := New(c)
	now := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v, err := s.Create(c.Targets[0])
	if err != nil {
		t.Fatal(err)
	}
	ip, _, err := s.Resolve(v.Host)
	if err != nil || ip.String() != c.EntryIP {
		t.Fatal(ip, err)
	}
	if _, _, err := s.Resolve(strings.Replace(v.Host, "7f000002", "7f000003", 1)); !errors.Is(err, ErrNotFound) {
		t.Fatal("forged target accepted")
	}
	armed, err := s.Arm(v.ID)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	arm2, _ := s.Arm(v.ID)
	if !arm2.RebindAt.Equal(armed.RebindAt) {
		t.Fatal("duplicate arm postponed switch")
	}
	now = armed.RebindAt
	ip, _, err = s.Resolve(strings.ToUpper(v.Host) + ".")
	if err != nil || ip.String() != c.Targets[0].IP {
		t.Fatal(ip, err)
	}
	if _, err := s.Create(c.Targets[0]); !errors.Is(err, ErrCapacity) {
		t.Fatal("capacity exceeded")
	}
	now = v.ExpiresAt
	if _, err := s.Get(v.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("expired session accessible")
	}
	if _, err := s.Create(c.Targets[0]); err != nil {
		t.Fatal("expired capacity not reclaimed", err)
	}
}
func TestConcurrentSessions(t *testing.T) {
	c := config.Defaults()
	s := New(c)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := s.Create(c.Targets[0])
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = s.Arm(v.ID)
			for j := 0; j < 20; j++ {
				_, _, _ = s.Resolve(v.Host)
				_, _ = s.Get(v.ID)
			}
		}()
	}
	wg.Wait()
}
func TestHexIPv6(t *testing.T) {
	if got := HexIP("::1"); got != "00000000000000000000000000000001" {
		t.Fatal(got)
	}
	if got := HexIP("::ffff:127.0.0.1"); got != "7f000001" {
		t.Fatal(got)
	}
}

func TestIPv6HostnameLengths(t *testing.T) {
	c := config.Defaults()
	c.EntryIP = "2001:db8::1"
	c.Targets[0].IP = "::1"
	s := New(c)
	v, err := s.Create(c.Targets[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Host) > 253 {
		t.Fatal("hostname too long", v.Host)
	}
	for _, label := range strings.Split(v.Host, ".") {
		if len(label) > 63 {
			t.Fatal("DNS label too long", label)
		}
	}
}
