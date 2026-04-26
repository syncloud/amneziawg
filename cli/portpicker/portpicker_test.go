package portpicker

import (
	"net"
	"testing"
)

func TestPickReturns55424WhenFree(t *testing.T) {
	port, err := Pick()
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if port != startPort {
		t.Errorf("expected %d when free, got %d", startPort, port)
	}
	if !isUDPFree(port) {
		t.Errorf("port %d returned as free but bind would fail", port)
	}
}

func TestPickWalksUpWhenStartTaken(t *testing.T) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: startPort})
	if err != nil {
		t.Skipf("could not occupy %d to set up test: %v", startPort, err)
	}
	defer conn.Close()

	port, err := Pick()
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if port == startPort {
		t.Errorf("expected port > %d when start is occupied, got %d", startPort, port)
	}
	if port < startPort || port > maxPort {
		t.Errorf("port %d out of [%d,%d]", port, startPort, maxPort)
	}
}
