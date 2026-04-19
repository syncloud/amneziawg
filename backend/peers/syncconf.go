package peers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

// serverTemplateContext mirrors the field shape the installer passes
// to config/awg-server.conf.tpl — same template is used at install
// time (with no peers) and at runtime (with the live peer set).
type serverTemplateContext struct {
	ServerPrivateKey string
	ListenPort       int
	Jc               int
	Jmin             int
	Jmax             int
	S1               int
	S2               int
	H1               uint32
	H2               uint32
	H3               uint32
	H4               uint32
	Peers            []serverPeer
}

type serverPeer struct {
	Name       string
	PublicKey  string
	AllowedIPs string
}

// syncServerConf rewrites the server awg0.conf from the current DB
// state, then hot-applies the peer section via awg syncconf.
//
// Strategy:
//   1. Re-render the full conf from config/awg-server.conf.tpl so a
//      daemon restart picks up the new peer set.
//   2. Shell out to awg-quick strip to get the syncconf-compatible
//      form, then pipe that to awg syncconf.
func (s *Service) syncServerConf() error {
	peers, err := s.DB.ListPeers()
	if err != nil {
		return err
	}

	ctx := serverTemplateContext{
		ServerPrivateKey: s.Config.PrivateKey,
		ListenPort:       s.Config.ListenPort,
		Jc:               s.Config.Obfuscation.Jc,
		Jmin:             s.Config.Obfuscation.Jmin,
		Jmax:             s.Config.Obfuscation.Jmax,
		S1:               s.Config.Obfuscation.S1,
		S2:               s.Config.Obfuscation.S2,
		H1:               s.Config.Obfuscation.H1,
		H2:               s.Config.Obfuscation.H2,
		H3:               s.Config.Obfuscation.H3,
		H4:               s.Config.Obfuscation.H4,
	}
	for _, p := range peers {
		ctx.Peers = append(ctx.Peers, serverPeer{
			Name:       p.Name,
			PublicKey:  p.PublicKey,
			AllowedIPs: p.AddressV4,
		})
	}

	var buf bytes.Buffer
	if err := s.ServerTemplate.Execute(&buf, ctx); err != nil {
		return fmt.Errorf("render server conf: %w", err)
	}
	if err := os.WriteFile(s.ServerConfPath, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("write %s: %w", s.ServerConfPath, err)
	}

	stripped, err := exec.Command(s.AwgQuickBinary, "strip", s.ServerConfPath).Output()
	if err != nil {
		return fmt.Errorf("awg-quick strip: %w", err)
	}
	return s.AWG.SyncConf(string(stripped))
}
