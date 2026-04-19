// Package portpicker picks a random high UDP port that is currently free
// on the host. Chosen at install time and persisted — other Syncloud
// apps may have grabbed a given port, so a static default is fragile.
//
// Port 51820 (the vanilla WireGuard default) is deliberately excluded:
// DPI systems flag it, and a Syncloud device behind TSPU/GFW should
// never use it.
package portpicker

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	minPort          = 10000
	maxPort          = 65535
	excludedWGPort   = 51820
	maxAttempts      = 100
)

// Pick returns a free random UDP port in [minPort, maxPort], excluding
// the vanilla WireGuard default.
func Pick() (int, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		port, err := randPort()
		if err != nil {
			return 0, err
		}
		if port == excludedWGPort {
			continue
		}
		if isUDPFree(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("could not find a free UDP port after %d attempts", maxAttempts)
}

func randPort() (int, error) {
	span := uint32(maxPort - minPort + 1)
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, err
	}
	n := binary.BigEndian.Uint32(buf[:]) % span
	return minPort + int(n), nil
}

func isUDPFree(port int) bool {
	addr := &net.UDPAddr{IP: net.IPv4zero, Port: port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
