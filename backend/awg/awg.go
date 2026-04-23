package awg

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Client struct {
	Binary    string
	Interface string
}

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
	LatestHandshake int64  `json:"latest_handshake"`
	RxBytes         int64  `json:"rx_bytes"`
	TxBytes         int64  `json:"tx_bytes"`
}

// awg show <iface> dump: first line is the interface, each subsequent
// line is a peer with tab-separated columns:
//   <pub> <preshared> <endpoint> <allowed_ips> <handshake> <rx> <tx> <keepalive>
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
	for _, line := range lines[1:] {
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
	var stderr strings.Builder
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}
