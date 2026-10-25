package peers

import (
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"backend/db"
)

func TestClientConfigDoesNotPointDNSAtTheTunnel(t *testing.T) {
	s, _ := newTestService(t)
	tpl, err := template.ParseFiles(filepath.Join("..", "..", "templates", "awg-client.conf.tpl"))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	s.ClientTemplate = tpl

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
	for _, line := range strings.Split(conf, "\n") {
		if strings.HasPrefix(line, "DNS") && strings.Contains(line, "10.9.0.1") {
			t.Fatalf("client DNS points at the tunnel address, where nothing answers DNS: %q", line)
		}
	}
}
