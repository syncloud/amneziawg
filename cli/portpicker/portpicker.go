package portpicker

import (
	"fmt"
	"net"
)

const (
	startPort = 55424
	maxPort   = 65535
)

func Pick() (int, error) {
	for port := startPort; port <= maxPort; port++ {
		if isUDPFree(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free UDP port in [%d,%d]", startPort, maxPort)
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
