package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/atemporalzen/horizon-winner/internal/config"
	"github.com/atemporalzen/horizon-winner/internal/dnsserver"
	"github.com/atemporalzen/horizon-winner/internal/httpserver"
	"github.com/atemporalzen/horizon-winner/internal/session"
	"github.com/miekg/dns"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("startup_failed", "error", err)
		os.Exit(1)
	}
}
func run() error {
	path := flag.String("config", "", "JSON config file (see configs/)")
	check := flag.Bool("check", false, "validate configuration and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	verbose := flag.Bool("verbose", false, "include DNS query events in JSON logs")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return nil
	}
	c, err := config.Load(*path)
	if err != nil {
		return err
	}
	if *check {
		fmt.Println("configuration valid")
		return nil
	}
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)
	token := os.Getenv("HORIZON_ADMIN_TOKEN")
	if token == "" {
		token, err = session.RandomToken()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Admin token (shown once): %s\n", token)
	}
	if len(token) < 24 || strings.ContainsAny(token, "\r\n") {
		return fmt.Errorf("HORIZON_ADMIN_TOKEN must be at least 24 characters without newlines")
	}
	store := session.New(c)
	handler := &dnsserver.Handler{Config: c, Store: store, Log: log}
	// Bind every listener before serving so partial startup cannot silently
	// leave a DNS or HTTP service running.
	httpListener, err := net.Listen("tcp", c.HTTPListen)
	if err != nil {
		return err
	}
	defer httpListener.Close()
	adminListener, err := net.Listen("tcp", c.AdminListen)
	if err != nil {
		return err
	}
	defer adminListener.Close()
	dnsUDP, err := net.ListenPacket("udp", c.DNSListen)
	if err != nil {
		return err
	}
	defer dnsUDP.Close()
	dnsTCP, err := net.Listen("tcp", c.DNSListen)
	if err != nil {
		return err
	}
	defer dnsTCP.Close()
	app := &httpserver.Server{Config: c, Store: store, Token: token, Log: log}
	httpServer := &http.Server{Handler: app.ExperimentHandler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 1 * time.Second, MaxHeaderBytes: 16384}
	adminServer := &http.Server{Handler: app.AdminHandler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, MaxHeaderBytes: 16384}
	httpServer.SetKeepAlivesEnabled(false)
	udp := &dns.Server{PacketConn: dnsUDP, Handler: handler, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second}
	tcp := &dns.Server{Listener: dnsTCP, Handler: handler, ReadTimeout: 2 * time.Second, WriteTimeout: 2 * time.Second, MaxTCPQueries: 20}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errors := make(chan error, 4)
	go func() { errors <- httpServer.Serve(httpListener) }()
	go func() { errors <- adminServer.Serve(adminListener) }()
	go func() { errors <- udp.ActivateAndServe() }()
	go func() { errors <- tcp.ActivateAndServe() }()
	log.Info("ready", "version", version, "control_url", c.ControlURL(), "http_listen", c.HTTPListen, "dns_listen", c.DNSListen)
	if entry := c.EntryIP; strings.HasPrefix(entry, "127.") || entry == "::1" {
		log.Info("local_demo", "note", "loopback entry does not cross a public-to-local LNA boundary; it validates application plumbing only")
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	var serveErr error
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case serveErr = <-errors:
			break loop
		case <-ticker.C:
			store.Sweep()
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	_ = adminServer.Shutdown(shutdownCtx)
	_ = udp.ShutdownContext(shutdownCtx)
	_ = tcp.ShutdownContext(shutdownCtx)
	if serveErr != nil && serveErr != http.ErrServerClosed {
		return serveErr
	}
	return nil
}
