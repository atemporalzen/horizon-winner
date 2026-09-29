package singularity

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHorizonTemplateLoadsOnlyDispatcher(t *testing.T) {
	w := httptest.NewRecorder()
	(&PayloadTemplateHandler{}).ServeHTTP(w, httptest.NewRequest("GET", "/soopayload.html", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "data-horizon-entry") || !strings.Contains(w.Body.String(), "/payload.js") || strings.Contains(w.Body.String(), "AwsMetadataExfil") {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestEntryErrorTimingMarker(t *testing.T) {
	h := DefaultHeadersHandler{NextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/missing", nil))
	if w.Code != 404 || !strings.Contains(w.Header().Get("Server-Timing"), "horizon-entry") {
		t.Fatal(w.Code, w.Header())
	}
}
func TestHorizonHexSessionProtocol(t *testing.T) {
	q, err := NewDNSQuery("s-cb00710a.7f000001-0123456789abcdef-fs-e.dynamic.example.com.")
	if err != nil || q.ResponseIPAddr != "203.0.113.10" || q.ResponseReboundIPAddr != "127.0.0.1" || q.DNSRebindingStrategy != "fs" {
		t.Fatal(q, err)
	}
}
