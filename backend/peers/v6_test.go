package peers

import (
	"net"
	"testing"
)

func TestPeerV6HostProducesValidIPv6(t *testing.T) {
	cases := []string{
		"10.9.0.2/32",
		"10.9.0.10/32",
		"10.9.0.255/32",
		"10.9.0.5",
	}
	for _, in := range cases {
		got := PeerV6Host(in)
		if got == "" {
			t.Fatalf("PeerV6Host(%q) returned empty", in)
		}
		ip := net.ParseIP(got)
		if ip == nil {
			t.Fatalf("PeerV6Host(%q) = %q, not parseable as IP", in, got)
		}
		if ip.To4() != nil {
			t.Fatalf("PeerV6Host(%q) = %q, parsed as v4", in, got)
		}
	}
}

func TestPeerV6HostUnique(t *testing.T) {
	a := PeerV6Host("10.9.0.2/32")
	b := PeerV6Host("10.9.0.3/32")
	if a == b || a == "" || b == "" {
		t.Fatalf("expected distinct non-empty, got %q vs %q", a, b)
	}
}
