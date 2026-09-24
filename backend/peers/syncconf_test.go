package peers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"backend/config"
	"backend/db"
)

func newTestService(t *testing.T) (*Service, string) {
	t.Helper()

	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	tpl, err := template.ParseFiles(filepath.Join("..", "..", "templates", "awg-server.conf.tpl"))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	confPath := filepath.Join(dir, "awg0.conf")
	return &Service{
		DB: database,
		Config: &config.Config{
			PrivateKey:  "cE7uWXcp6JQYr1Jnq0yrqyHpEGk8p+Uc9MKv1Kf2a1g=",
			ListenPort:  55424,
			Obfuscation: config.Obfuscation{Jc: 4, Jmin: 50, Jmax: 1000, S1: 36, S2: 130, H1: 1, H2: 2, H3: 3, H4: 4},
		},
		ServerTemplate: tpl,
		ServerConfPath: confPath,
		Subnet:         "10.9.0.0/24",
	}, confPath
}

func TestRenderServerConfRestoresPeersOverPeerlessFile(t *testing.T) {
	s, confPath := newTestService(t)

	if _, err := s.DB.InsertPeer(db.Peer{
		Name:       "phone",
		PublicKey:  "1Le9eNZ1ghxociDc+cDaTpW/OiRmHuDXiJM4Bhiy/2c=",
		PrivateKey: "OMPz8nCPXkqAr6cyQ0XaTDEwyoy5hK8Nmbq3fFqO5Wo=",
		AddressV4:  "10.9.0.2/32",
	}); err != nil {
		t.Fatalf("insert peer: %v", err)
	}

	if err := os.WriteFile(confPath, []byte("[Interface]\nListenPort = 55424\n"), 0600); err != nil {
		t.Fatalf("write peerless conf: %v", err)
	}

	if err := s.RenderServerConf(); err != nil {
		t.Fatalf("render: %v", err)
	}

	out, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read conf: %v", err)
	}
	conf := string(out)

	for _, want := range []string{
		"[Peer]",
		"PublicKey  = 1Le9eNZ1ghxociDc+cDaTpW/OiRmHuDXiJM4Bhiy/2c=",
		"AllowedIPs = 10.9.0.2/32, 2001:db8:a7c9::2/128",
		"ListenPort = 55424",
	} {
		if !strings.Contains(conf, want) {
			t.Fatalf("rendered conf missing %q:\n%s", want, conf)
		}
	}
}

func TestRenderServerConfWithoutPeers(t *testing.T) {
	s, confPath := newTestService(t)

	if err := s.RenderServerConf(); err != nil {
		t.Fatalf("render: %v", err)
	}

	out, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("read conf: %v", err)
	}
	if strings.Contains(string(out), "[Peer]") {
		t.Fatalf("expected no peer sections:\n%s", out)
	}
}
