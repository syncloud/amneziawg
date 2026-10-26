package peers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"backend/db"
)

func renderClientConfig(t *testing.T, resolvConf string) string {
	t.Helper()
	s, _ := newTestService(t)
	tpl, err := template.ParseFiles(filepath.Join("..", "..", "templates", "awg-client.conf.tpl"))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	s.ClientTemplate = tpl
	if resolvConf != "" {
		if err := os.WriteFile(s.LanResolver.ResolvConf, []byte(resolvConf), 0644); err != nil {
			t.Fatalf("write resolv.conf: %v", err)
		}
	}

	id, err := s.DB.InsertPeer(db.Peer{
		Name:       "phone",
		PublicKey:  "1Le9eNZ1ghxociDc+cDaTpW/OiRmHuDXiJM4Bhiy/2c=",
		PrivateKey: "OMPz8nCPXkqAr6cyQ0XaTDEwyoy5hK8Nmbq3fFqO5Wo=",
		AddressV4:  "10.9.0.2/32",
	})
	if err != nil {
		t.Fatalf("insert peer: %v", err)
	}

	conf, _, err := s.ClientConfig(id)
	if err != nil {
		t.Fatalf("client config: %v", err)
	}
	return conf
}

func TestClientConfigPushesLanResolver(t *testing.T) {
	conf := renderClientConfig(t, "domain home.arpa\nsearch home.arpa\nnameserver 192.168.1.1\n")
	want := "DNS        = 192.168.1.1, home.arpa\n"
	if !strings.Contains(conf, want) {
		t.Fatalf("client config missing %q:\n%s", want, conf)
	}
}

func TestClientConfigFallsBackToPublicResolver(t *testing.T) {
	conf := renderClientConfig(t, "nameserver 127.0.0.11\n")
	want := "DNS        = 8.8.8.8, 2001:4860:4860::8888\n"
	if !strings.Contains(conf, want) {
		t.Fatalf("client config missing %q:\n%s", want, conf)
	}
}

func TestClientConfigDoesNotPointDNSAtTheTunnel(t *testing.T) {
	conf := renderClientConfig(t, "")
	for _, line := range strings.Split(conf, "\n") {
		if strings.HasPrefix(line, "DNS") && strings.Contains(line, "10.9.0.1") {
			t.Fatalf("client DNS points at the tunnel address, where nothing answers DNS: %q", line)
		}
	}
}
