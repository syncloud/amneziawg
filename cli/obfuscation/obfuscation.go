package obfuscation

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

type Params struct {
	Jc   int
	Jmin int
	Jmax int
	S1   int
	S2   int
	H1   uint32
	H2   uint32
	H3   uint32
	H4   uint32
}

func Generate() (Params, error) {
	p := Params{
		Jmin: 50,
		Jmax: 1000,
	}

	jc, err := randIntRange(3, 10)
	if err != nil {
		return p, err
	}
	p.Jc = jc

	s1, err := randIntRange(15, 150)
	if err != nil {
		return p, err
	}
	p.S1 = s1

	// AmneziaWG: S2 must not equal S1+56.
	for {
		s2, err := randIntRange(15, 150)
		if err != nil {
			return p, err
		}
		if s2 != s1+56 {
			p.S2 = s2
			break
		}
	}

	// AmneziaWG: H1..H4 must be distinct and not in {1,2,3,4} (reserved WG message types).
	seen := map[uint32]bool{1: true, 2: true, 3: true, 4: true}
	var hs [4]uint32
	for i := 0; i < 4; i++ {
		for {
			h, err := randUint32GE(5)
			if err != nil {
				return p, err
			}
			if !seen[h] {
				seen[h] = true
				hs[i] = h
				break
			}
		}
	}
	p.H1, p.H2, p.H3, p.H4 = hs[0], hs[1], hs[2], hs[3]

	return p, nil
}

func randIntRange(lo, hi int) (int, error) {
	if hi < lo {
		return 0, fmt.Errorf("range hi=%d < lo=%d", hi, lo)
	}
	span := uint32(hi - lo + 1)
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, err
	}
	n := binary.BigEndian.Uint32(buf[:]) % span
	return lo + int(n), nil
}

func randUint32GE(min uint32) (uint32, error) {
	var buf [4]byte
	for {
		if _, err := rand.Read(buf[:]); err != nil {
			return 0, err
		}
		n := binary.BigEndian.Uint32(buf[:])
		if n >= min {
			return n, nil
		}
	}
}
