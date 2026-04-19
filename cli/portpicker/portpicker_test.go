package portpicker

import "testing"

func TestPick_InRangeAndFree(t *testing.T) {
	for i := 0; i < 5; i++ {
		port, err := Pick()
		if err != nil {
			t.Fatalf("Pick() error: %v", err)
		}
		if port < minPort || port > maxPort {
			t.Errorf("port %d out of [%d,%d]", port, minPort, maxPort)
		}
		if port == excludedWGPort {
			t.Errorf("got excluded WG port %d", port)
		}
		if !isUDPFree(port) {
			t.Errorf("port %d returned as free but bind would fail", port)
		}
	}
}

func TestIsUDPFree(t *testing.T) {
	port, err := Pick()
	if err != nil {
		t.Fatal(err)
	}
	if !isUDPFree(port) {
		t.Errorf("fresh port %d should be free", port)
	}
}
