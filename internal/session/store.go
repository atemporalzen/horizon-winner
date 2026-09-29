package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/atemporalzen/horizon-winner/internal/config"
)

var ErrNotFound = errors.New("session missing or expired")
var ErrCapacity = errors.New("session capacity reached")

type Session struct {
	ID        string        `json:"id"`
	Host      string        `json:"host"`
	URL       string        `json:"url"`
	RunKey    string        `json:"-"`
	Target    config.Target `json:"target"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	RebindAt  time.Time     `json:"rebind_at,omitempty"`
	Queries   uint64        `json:"queries"`
}

type Store struct {
	mu       sync.Mutex
	sessions map[string]Session
	c        config.Config
	now      func() time.Time
}

func New(c config.Config) *Store { return &Store{sessions: map[string]Session{}, c: c, now: time.Now} }
func RandomToken() (string, error) {
	b := make([]byte, 24)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
func HexIP(ip string) string {
	a := netip.MustParseAddr(ip).Unmap()
	return hex.EncodeToString(a.AsSlice())
}

func (s *Store) sweepLocked(now time.Time) {
	for id, v := range s.sessions {
		if !now.Before(v.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}
func (s *Store) Sweep() { s.mu.Lock(); defer s.mu.Unlock(); s.sweepLocked(s.now()) }

func (s *Store) Create(t config.Target) (Session, error) {
	id, err := RandomToken()
	if err != nil {
		return Session{}, err
	}
	// 96-bit IDs leave the target+session IPv6 DNS label within 63 bytes.
	id = id[:24]
	key, err := RandomToken()
	if err != nil {
		return Session{}, err
	}
	now := s.now()
	host := "s-" + HexIP(s.c.EntryIP) + "." + HexIP(t.IP) + "-" + id + "-fs-e." + s.c.Domain
	v := Session{ID: id, Host: host, URL: "http://" + s.c.Host(host) + "/run", RunKey: key, Target: t, CreatedAt: now, ExpiresAt: now.Add(time.Duration(s.c.SessionTTLSeconds) * time.Second)}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepLocked(now)
	if len(s.sessions) >= s.c.MaxSessions {
		return Session{}, ErrCapacity
	}
	s.sessions[id] = v
	return v, nil
}

func (s *Store) getLocked(id string) (Session, error) {
	v, ok := s.sessions[id]
	if !ok || !s.now().Before(v.ExpiresAt) {
		delete(s.sessions, id)
		return Session{}, ErrNotFound
	}
	return v, nil
}
func (s *Store) Get(id string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getLocked(id)
}
func (s *Store) ByHost(host string) (Session, error) {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	s.mu.Lock()
	defer s.mu.Unlock()
	// Lookups bind the complete hostname to server-selected addresses. The
	// encoded IPs in the label never authorize a new target.
	for id, v := range s.sessions {
		if v.Host == host {
			return s.getLocked(id)
		}
	}
	return Session{}, ErrNotFound
}
func (s *Store) Arm(id string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, err := s.getLocked(id)
	if err != nil {
		return v, err
	}
	if v.RebindAt.IsZero() {
		v.RebindAt = s.now().Add(time.Duration(s.c.RebindAfterSeconds) * time.Second)
		s.sessions[id] = v
	}
	return v, nil
}
func (s *Store) Resolve(host string) (netip.Addr, Session, error) {
	v, err := s.ByHost(host)
	if err != nil {
		return netip.Addr{}, v, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, err = s.getLocked(v.ID)
	if err != nil {
		return netip.Addr{}, v, err
	}
	v.Queries++
	s.sessions[v.ID] = v
	ip := s.c.EntryIP
	if !v.RebindAt.IsZero() && !s.now().Before(v.RebindAt) {
		ip = v.Target.IP
	}
	return netip.MustParseAddr(ip).Unmap(), v, nil
}
