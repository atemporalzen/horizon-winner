package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// The fixture contains synthetic data, with no metadata, credentials or files.
func main() {
	listen := flag.String("listen", "127.0.0.2:8080", "fixture listener")
	marker := flag.String("marker", "horizon-lab-proof-v1", "synthetic proof marker")
	allowed := flag.String("allowed-hosts", "", "exact comma-separated Host headers; empty demonstrates a vulnerable service")
	coop := flag.Bool("sever-opener", false, "send COOP same-origin to exercise opener loss")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if *coop {
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		}
		if *allowed != "" {
			valid := false
			for _, h := range strings.Split(*allowed, ",") {
				if r.Host == strings.TrimSpace(h) {
					valid = true
				}
			}
			if !valid {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(421)
				fmt.Fprint(w, "horizon-host-denied: unexpected Host")
				return
			}
		}
		if r.URL.Path != "/proof" {
			http.NotFound(w, r)
			return
		}
		if r.Method != "GET" {
			http.Error(w, "method not allowed", 405)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<!doctype html><html><head><title>Horizon synthetic target</title></head><body><h1>Mock target reached</h1><pre>%s</pre><p>Observed Host: %s</p><p>This response contains synthetic test data.</p></body></html>", template.HTMLEscapeString(*marker), template.HTMLEscapeString(r.Host))
	})
	s := &http.Server{Addr: *listen, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	s.SetKeepAlivesEnabled(false)
	log.Printf("Synthetic fixture on %s; Host validation enabled=%v", *listen, *allowed != "")
	log.Fatal(s.ListenAndServe())
}
