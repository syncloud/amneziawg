// Package awg wraps the bundled awg CLI for keypair generation, live
// status queries, and hot-applying peer changes via awg syncconf.
package awg

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Client struct {
	Binary    string // path to the awg binary
	Interface string // e.g. "awg0"
}

// GenerateKeypair returns (privateKey, publicKey) for a fresh client.
// We shell out to awg genkey / awg pubkey rather than re-implementing
// Curve25519 — the upstream binary is the source of truth.
func (c *Client) GenerateKeypair() (string, string, error) {
	priv, err := run(c.Binary, "genkey")
	if err != nil {
		return "", "", fmt.Errorf("awg genkey: %w", err)
	}
	priv = strings.TrimSpace(priv)

	cmd := exec.Command(c.Binary, "pubkey")
	cmd.Stdin = strings.NewReader(priv + "\n")
	out, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("awg pubkey: %w", err)
	}
	pub := strings.TrimSpace(string(out))
	return priv, pub, nil
}

type PeerStatus struct {
	PublicKey       string `json:"public_key"`
	Endpoint        string `json:"endpoint"`
	AllowedIPs      string `json:"allowed_ips"`
	LatestHandshake int64  `json:"latest_handshake"` // unix seconds
	RxBytes         int64  `json:"rx_bytes"`
	TxBytes         int64  `json:"tx_bytes"`
}

// Dump parses `awg show <iface> dump`. First line is the interface
// itself; remaining lines are peers. Columns are tab-separated.
//
// Peer line layout: <pub_key> <preshared_key> <endpoint> <allowed_ips>
//   <latest_handshake> <rx_bytes> <tx_bytes> <persistent_keepalive>
func (c *Client) Dump() ([]PeerStatus, error) {
	out, err := run(c.Binary, "show", c.Interface, "dump")
	if err != nil {
		return nil, fmt.Errorf("awg show dump: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 1 {
		return nil, nil
	}
	var peers []PeerStatus
	for _, line := range lines[1:] { // skip interface line
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}
		p := PeerStatus{
			PublicKey:  fields[0],
			Endpoint:   fields[2],
			AllowedIPs: fields[3],
		}
		p.LatestHandshake, _ = strconv.ParseInt(fields[4], 10, 64)
		p.RxBytes, _ = strconv.ParseInt(fields[5], 10, 64)
		p.TxBytes, _ = strconv.ParseInt(fields[6], 10, 64)
		peers = append(peers, p)
	}
	return peers, nil
}

// SyncConf hot-applies the conf file to a running interface without
// dropping peers (equivalent of: awg syncconf <iface> <(awg-quick strip conf)).
func (c *Client) SyncConf(stripped string) error {
	cmd := exec.Command(c.Binary, "syncconf", c.Interface, "/dev/stdin")
	cmd.Stdin = strings.NewReader(stripped)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("awg syncconf: %w (%s)", err, string(out))
	}
	return nil
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
