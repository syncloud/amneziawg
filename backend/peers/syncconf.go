package peers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

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
