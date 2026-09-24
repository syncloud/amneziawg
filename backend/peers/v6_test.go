package peers

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestPeerV6HostIsGloballyScoped(t *testing.T) {
	global := &net.IPNet{IP: net.ParseIP("2000::"), Mask: net.CIDRMask(3, 128)}
	ula := &net.IPNet{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)}

	for _, in := range []string{"10.9.0.2/32", "10.9.0.200/32"} {
		ip := net.ParseIP(PeerV6Host(in))
		require.NotNil(t, ip, "PeerV6Host(%q) not parseable", in)
		require.False(t, ula.Contains(ip),
			"PeerV6Host(%q) = %s is a ULA: RFC 6724 labels it 13, so Android refuses it "+
				"as a source for public destinations and silently falls back to IPv4", in, ip)
		require.True(t, global.Contains(ip),
			"PeerV6Host(%q) = %s is outside 2000::/3, so it will not share RFC 6724 "+
				"label 1 with public destinations", in, ip)
	}
}

func TestPeerV6HostUnique(t *testing.T) {
	a := PeerV6Host("10.9.0.2/32")
	b := PeerV6Host("10.9.0.3/32")
	if a == b || a == "" || b == "" {
		t.Fatalf("expected distinct non-empty, got %q vs %q", a, b)
	}
}
