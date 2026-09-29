package dnsserver

import (
	"net"
	"testing"

	"github.com/atemporalzen/horizon-winner/internal/config"
	"github.com/atemporalzen/horizon-winner/internal/session"
	"github.com/miekg/dns"
)

type recorder struct{ m *dns.Msg }

func (r *recorder) LocalAddr() net.Addr         { return &net.UDPAddr{} }
func (r *recorder) RemoteAddr() net.Addr        { return &net.UDPAddr{} }
func (r *recorder) WriteMsg(m *dns.Msg) error   { r.m = m; return nil }
func (r *recorder) Write(p []byte) (int, error) { return len(p), nil }
func (r *recorder) Close() error                { return nil }
func (r *recorder) TsigStatus() error           { return nil }
func (r *recorder) TsigTimersOnly(bool)         {}
func (r *recorder) Hijack()                     {}

func TestAuthorityAndAddressFamilies(t *testing.T) {
	c := config.Defaults()
	s := session.New(c)
	v, _ := s.Create(c.Targets[0])
	h := &Handler{Config: c, Store: s}
	for _, test := range []struct {
		name          string
		typ           uint16
		code, answers int
	}{
		{v.Host, dns.TypeA, 0, 1}, {v.Host, dns.TypeAAAA, 0, 0}, {"control.rebind.test", dns.TypeA, 0, 1},
		{"s-forged.rebind.test", dns.TypeA, dns.RcodeNameError, 0}, {"other.test", dns.TypeA, dns.RcodeRefused, 0},
		{"rebind.test", dns.TypeNS, 0, 1}, {"rebind.test", dns.TypeSOA, 0, 1}, {"xcontrol.rebind.test", dns.TypeA, dns.RcodeNameError, 0},
	} {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(test.name), test.typ)
		r := &recorder{}
		h.ServeDNS(r, m)
		if r.m.Rcode != test.code || len(r.m.Answer) != test.answers {
			t.Errorf("%s %d: %#v", test.name, test.typ, r.m)
		}
		if r.m.RecursionAvailable {
			t.Fatal("open recursion advertised")
		}
	}
	c.EntryIP = "2001:db8::1"
	c.Targets[0].IP = "::1"
	s = session.New(c)
	v, _ = s.Create(c.Targets[0])
	h = &Handler{Config: c, Store: s}
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(v.Host), dns.TypeAAAA)
	r := &recorder{}
	h.ServeDNS(r, m)
	if len(r.m.Answer) != 1 {
		t.Fatal("IPv6 answer missing")
	}
}
func TestUDPAndTCP(t *testing.T) {
	c := config.Defaults()
	h := &Handler{Config: c, Store: session.New(c)}
	for _, network := range []string{"udp", "tcp"} {
		t.Run(network, func(t *testing.T) {
			ready := make(chan struct{})
			srv := &dns.Server{Handler: h, NotifyStartedFunc: func() { close(ready) }}
			var addr string
			if network == "udp" {
				conn, err := net.ListenPacket("udp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				srv.PacketConn = conn
				addr = conn.LocalAddr().String()
			} else {
				l, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				srv.Listener = l
				addr = l.Addr().String()
			}
			done := make(chan error, 1)
			go func() { done <- srv.ActivateAndServe() }()
			<-ready
			defer func() {
				_ = srv.Shutdown()
				if err := <-done; err != nil {
					t.Error(err)
				}
			}()
			m := new(dns.Msg)
			m.SetQuestion("control.rebind.test.", dns.TypeA)
			response, _, err := (&dns.Client{Net: network}).Exchange(m, addr)
			if err != nil || len(response.Answer) != 1 {
				t.Fatal(response, err)
			}
		})
	}
}
