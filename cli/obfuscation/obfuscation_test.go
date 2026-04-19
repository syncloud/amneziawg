package obfuscation

import "testing"

func TestGenerate_Constraints(t *testing.T) {
	// Run the generator many times to exercise the random space.
	for i := 0; i < 1000; i++ {
		p, err := Generate()
		if err != nil {
			t.Fatalf("Generate() error: %v", err)
		}

		if p.Jc < 3 || p.Jc > 10 {
			t.Errorf("Jc=%d out of [3,10]", p.Jc)
		}
		if p.Jmin != 50 {
			t.Errorf("Jmin=%d, want 50", p.Jmin)
		}
		if p.Jmax != 1000 {
			t.Errorf("Jmax=%d, want 1000", p.Jmax)
		}
		if p.S1 < 15 || p.S1 > 150 {
			t.Errorf("S1=%d out of [15,150]", p.S1)
		}
		if p.S2 < 15 || p.S2 > 150 {
			t.Errorf("S2=%d out of [15,150]", p.S2)
		}
		if p.S2 == p.S1+56 {
			t.Errorf("S2 (%d) must not equal S1+56 (%d)", p.S2, p.S1+56)
		}

		hs := []uint32{p.H1, p.H2, p.H3, p.H4}
		seen := map[uint32]bool{}
		for _, h := range hs {
			if h < 5 {
				t.Errorf("H value %d < 5", h)
			}
			if h == 1 || h == 2 || h == 3 || h == 4 {
				t.Errorf("H value %d is reserved WG message type", h)
			}
			if seen[h] {
				t.Errorf("H values not distinct: %d repeated", h)
			}
			seen[h] = true
		}
	}
}

func TestRandIntRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		n, err := randIntRange(5, 10)
		if err != nil {
			t.Fatal(err)
		}
		if n < 5 || n > 10 {
			t.Errorf("randIntRange(5,10)=%d out of range", n)
		}
	}
}

func TestRandIntRange_SingleValue(t *testing.T) {
	n, err := randIntRange(7, 7)
	if err != nil {
		t.Fatal(err)
	}
	if n != 7 {
		t.Errorf("randIntRange(7,7)=%d, want 7", n)
	}
}

func TestRandIntRange_Invalid(t *testing.T) {
	if _, err := randIntRange(10, 5); err == nil {
		t.Error("expected error for hi < lo")
	}
}
