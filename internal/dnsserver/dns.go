package dnsserver

import (
	"log/slog"
	"net"
	"net/netip"
	"strings"

	"github.com/atemporalzen/horizon-winner/internal/config"
	"github.com/atemporalzen/horizon-winner/internal/session"
	"github.com/miekg/dns"
)

type Handler struct {
	Config config.Config
	Store  *session.Store
	Log    *slog.Logger
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = false
	if r.Opcode != dns.OpcodeQuery || len(r.Question) != 1 {
		m.Rcode = dns.RcodeFormatError
		_ = w.WriteMsg(m)
		return
	}
	q := r.Question[0]
	name := strings.ToLower(q.Name)
	zone := dns.Fqdn(h.Config.Domain)
	if !dns.IsSubDomain(zone, name) || q.Qclass != dns.ClassINET {
		m.Rcode = dns.RcodeRefused
		_ = w.WriteMsg(m)
		return
	}
	m.Authoritative = true
	header := dns.RR_Header{Name: q.Name, Rrtype: q.Qtype, Class: dns.ClassINET, Ttl: 0}
	if name == zone && (q.Qtype == dns.TypeNS || q.Qtype == dns.TypeSOA) {
		if q.Qtype == dns.TypeNS {
			m.Answer = []dns.RR{&dns.NS{Hdr: header, Ns: dns.Fqdn(h.Config.DNSNameserver)}}
		} else {
			m.Answer = []dns.RR{h.soa(q.Name)}
		}
		_ = w.WriteMsg(m)
		return
	}
	ip := netip.MustParseAddr(h.Config.EntryIP).Unmap()
	stable := name == zone || name == dns.Fqdn("control."+h.Config.Domain) || name == dns.Fqdn(h.Config.DNSNameserver)
	if !stable {
		var err error
		var v session.Session
		ip, v, err = h.Store.Resolve(name)
		if err != nil {
			m.Rcode = dns.RcodeNameError
			m.Ns = []dns.RR{h.soa(zone)}
			_ = w.WriteMsg(m)
			return
		}
		if h.Log != nil {
			h.Log.Debug("dns_answer", "session", v.ID, "query_type", q.Qtype, "answer", ip.String(), "queries", v.Queries)
		}
	}
	switch {
	case q.Qtype == dns.TypeA && ip.Is4():
		m.Answer = []dns.RR{&dns.A{Hdr: header, A: net.IP(ip.AsSlice())}}
	case q.Qtype == dns.TypeAAAA && ip.Is6():
		m.Answer = []dns.RR{&dns.AAAA{Hdr: header, AAAA: net.IP(ip.AsSlice())}}
	default:
		m.Ns = []dns.RR{h.soa(zone)} // NODATA for the other family; no phase change.
	}
	_ = w.WriteMsg(m)
}
func (h *Handler) soa(name string) *dns.SOA {
	return &dns.SOA{Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 0}, Ns: dns.Fqdn(h.Config.DNSNameserver), Mbox: dns.Fqdn(h.Config.DNSContact), Serial: 1, Refresh: 60, Retry: 60, Expire: 300, Minttl: 0}
}
