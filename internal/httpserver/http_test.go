package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atemporalzen/horizon-winner/internal/config"
	"github.com/atemporalzen/horizon-winner/internal/session"
)

func TestAuthIsolationAndRunLifecycle(t *testing.T) {
	c := config.Defaults()
	s := &Server{Config: c, Store: session.New(c), Token: "test-token-01234567890123456789"}
	request := func(host, method, path, body, token, origin, key string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://"+host+path, strings.NewReader(body))
		r.Host = host
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Horizon-Run", key)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	if w := request("untrusted.test:8080", "GET", "/", "", "", "", ""); w.Code != 421 {
		t.Fatal(w.Code)
	}
	if w := request(c.ControlHost(), "POST", "/api/sessions", `{"target_id":"lab"}`, "", "", ""); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := request(c.ControlHost(), "POST", "/api/sessions", `{"target_id":"lab"}`, s.Token, "http://evil.test", ""); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if w := request(c.ControlHost(), "POST", "/api/sessions", `{"target_id":"lab","ip":"10.0.0.1"}`, s.Token, "", ""); w.Code != 400 {
		t.Fatal("arbitrary target accepted", w.Code)
	}
	w := request(c.ControlHost(), "POST", "/api/sessions", `{"target_id":"lab"}`, s.Token, c.ControlURL(), "")
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var v session.Session
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	stored, _ := s.Store.Get(v.ID)
	if bytes.Contains(w.Body.Bytes(), []byte(stored.RunKey)) {
		t.Fatal("run key leaked in control response")
	}
	host := c.Host(v.Host)
	if w := request(host, "GET", "/run", "", "", "", ""); w.Code != 200 || w.Header().Get("Cross-Origin-Opener-Policy") != "" {
		t.Fatal("runner opener broken", w.Code)
	}
	if w := request(host, "POST", "/api/sessions", `{"target_id":"lab"}`, s.Token, "", ""); w.Code != 405 {
		t.Fatal("control API exposed to runner", w.Code)
	}
	if w := request(host, "POST", "/api/arm", "", "", "", ""); w.Code != 401 {
		t.Fatal(w.Code)
	}
	if w := request(host, "POST", "/api/arm", "", "", "http://evil.test", stored.RunKey); w.Code != 403 {
		t.Fatal(w.Code)
	}
	if w := request(host, "POST", "/api/arm", "", "", "http://"+host, stored.RunKey); w.Code != 200 {
		t.Fatal(w.Code)
	}
	armed, _ := s.Store.Get(v.ID)
	if armed.RebindAt.IsZero() {
		t.Fatal("session not armed")
	}
	if w := request(host, "GET", "/proof", "", "", "", ""); w.Code != 200 || !strings.Contains(w.Body.String(), "data-horizon-entry") {
		t.Fatal("entry lacks marker")
	}
}

func TestEmbeddedAssetsServeFromAnyDirectory(t *testing.T) {
	c := config.Defaults()
	s := &Server{Config: c, Store: session.New(c), Token: "test"}
	for _, path := range []string{"/", "/controller.js", "/style.css"} {
		r := httptest.NewRequest(http.MethodGet, "http://"+c.ControlHost()+path, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatal(path, w.Code)
		}
	}
}

func TestListenerIsolation(t *testing.T) {
	c := config.Defaults()
	s := &Server{Config: c, Store: session.New(c), Token: "test"}
	r := httptest.NewRequest("GET", c.ControlURL()+"/api/config", nil)
	r.Header.Set("Authorization", "Bearer test")
	w := httptest.NewRecorder()
	s.ExperimentHandler().ServeHTTP(w, r)
	if w.Code != 421 {
		t.Fatal("public listener accepted admin Host", w.Code)
	}
	r = httptest.NewRequest("GET", "http://other.test:9090/", nil)
	w = httptest.NewRecorder()
	s.AdminHandler().ServeHTTP(w, r)
	if w.Code != 421 {
		t.Fatal("admin listener accepted foreign Host", w.Code)
	}
}
