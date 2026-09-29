package httpserver

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/atemporalzen/horizon-winner/internal/config"
	"github.com/atemporalzen/horizon-winner/internal/session"
	"github.com/atemporalzen/horizon-winner/web"
)

type Server struct {
	Config config.Config
	Store  *session.Store
	Token  string
	Log    *slog.Logger
}

// Separate handlers prevent a crafted public Host header from reaching the
// admin API through the experiment listener.
func (s *Server) AdminHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Host) != s.Config.ControlHost() {
			http.Error(w, "unrecognized admin Host", 421)
			return
		}
		s.ServeHTTP(w, r)
	})
}
func (s *Server) ExperimentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Host) == s.Config.ControlHost() {
			http.Error(w, "admin API unavailable on experiment listener", 421)
			return
		}
		s.ServeHTTP(w, r)
	})
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Horizon-Origin", "entry")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
	// Runner and popup deliberately retain their same-origin opener reference.
	// COOP/noopener on those pages would invalidate the experiment itself.
	host := strings.ToLower(r.Host)
	if host == s.Config.ControlHost() {
		s.control(w, r)
		return
	}
	name := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		name = h
	}
	v, err := s.Store.ByHost(name)
	if err != nil || host != s.Config.Host(v.Host) {
		http.Error(w, "unrecognized Host", http.StatusMisdirectedRequest)
		return
	}
	s.run(w, r, v)
}

func (s *Server) asset(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", 405)
		return
	}
	f, err := web.Files.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	data, _ := io.ReadAll(f)
	if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	}
	if strings.HasSuffix(path, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	if r.Method != "HEAD" {
		_, _ = w.Write(data)
	}
}

func (s *Server) control(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	switch r.URL.Path {
	case "/":
		s.asset(w, r, "index.html")
	case "/style.css":
		s.asset(w, r, "style.css")
	case "/controller.js":
		s.asset(w, r, "controller.js")
	case "/healthz":
		if r.Method != "GET" {
			http.Error(w, "method not allowed", 405)
			return
		}
		writeJSON(w, 200, map[string]any{"status": "ok", "service": "horizon-winner"})
	case "/api/config":
		if r.Method != "GET" {
			http.Error(w, "method not allowed", 405)
			return
		}
		if !s.auth(w, r) {
			return
		}
		writeJSON(w, 200, map[string]any{"targets": s.Config.Targets, "control_url": s.Config.ControlURL(), "rebind_after_seconds": s.Config.RebindAfterSeconds})
	case "/api/sessions":
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", 405)
			return
		}
		if !s.auth(w, r) || !sameOrigin(w, r) {
			return
		}
		var body struct {
			TargetID string `json:"target_id"`
		}
		if err := decode(w, r, &body); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		t, ok := s.Config.Target(body.TargetID)
		if !ok {
			http.Error(w, "unknown target", 400)
			return
		}
		v, err := s.Store.Create(t)
		if errors.Is(err, session.ErrCapacity) {
			http.Error(w, err.Error(), 429)
			return
		}
		if err != nil {
			http.Error(w, "could not create session", 500)
			return
		}
		if s.Log != nil {
			s.Log.Info("session_created", "session", v.ID, "target", v.Target.ID)
		}
		writeJSON(w, 201, v)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) run(w http.ResponseWriter, r *http.Request, v session.Session) {
	switch r.URL.Path {
	case "/run":
		s.asset(w, r, "run.html")
	case "/runner.js", "/engine.js", "/style.css":
		s.asset(w, r, strings.TrimPrefix(r.URL.Path, "/"))
	case "/popup":
		s.asset(w, r, "popup.html")
	case "/api/run":
		if r.Method != "GET" {
			http.Error(w, "method not allowed", 405)
			return
		}
		writeJSON(w, 200, map[string]any{"session": v, "run_key": v.RunKey, "entry_ip": s.Config.EntryIP, "poll_interval_ms": s.Config.PollIntervalMS, "request_timeout_ms": s.Config.RequestTimeoutMS, "run_timeout_seconds": s.Config.RunTimeoutSeconds, "max_response_bytes": s.Config.MaxResponseBytes, "rebind_after_seconds": s.Config.RebindAfterSeconds})
	case "/api/arm":
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", 405)
			return
		}
		if !sameOrigin(w, r) {
			return
		}
		if !equal(r.Header.Get("X-Horizon-Run"), v.RunKey) {
			http.Error(w, "invalid session capability", 401)
			return
		}
		armed, err := s.Store.Arm(v.ID)
		if err != nil {
			http.Error(w, "session expired", 410)
			return
		}
		if s.Log != nil {
			s.Log.Info("session_armed", "session", v.ID, "rebind_at", armed.RebindAt)
		}
		writeJSON(w, 200, armed)
	default:
		// A target-path request still reaching the entry service is identifiable.
		// Report it as entry content even when the configured target path collides
		// with no route; never fabricate a target proof.
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "method not allowed", 405)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if r.Method != "HEAD" {
			_, _ = io.WriteString(w, "<!doctype html><html><head><title>Entry service</title></head><body data-horizon-entry>Horizon entry service. DNS has not switched for this connection.</body></html>")
		}
	}
}

func equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func (s *Server) auth(w http.ResponseWriter, r *http.Request) bool {
	if !equal(r.Header.Get("Authorization"), "Bearer "+s.Token) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "authorization required", 401)
		return false
	}
	return true
}
func sameOrigin(w http.ResponseWriter, r *http.Request) bool {
	// Command-line API clients may omit Origin. Browser requests must match.
	if o := r.Header.Get("Origin"); o != "" && o != "http://"+r.Host {
		http.Error(w, "origin rejected", 403)
		return false
	}
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
		http.Error(w, "request site rejected", 403)
		return false
	}
	return true
}
func decode(w http.ResponseWriter, r *http.Request, v any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("JSON required")
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
