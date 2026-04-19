package peers

import (
	"bytes"
	"fmt"
	"net"
	"text/template"

	"github.com/skip2/go-qrcode"

	"backend/awg"
	"backend/config"
	"backend/db"
)

// Service owns peer lifecycle: key generation, address allocation,
// client-config rendering, hot-apply to the running awg interface.
//
// All templates are loaded at construction from the config dir; the
// service never reads template strings from Go source.
type Service struct {
	DB     *db.DB
	AWG    *awg.Client
	Config *config.Config

	ServerTemplate *template.Template
	ClientTemplate *template.Template

	// ServerConfPath is where the rendered awg0.conf lives — rewritten
	// on every peer mutation so a daemon restart is also safe.
	ServerConfPath string

	// AwgQuickBinary is the path to the bundled awg-quick; used to
	// `strip` the server conf before feeding it to `awg syncconf`.
	AwgQuickBinary string

	// Subnet is the pool /24 — server is .1, peers start at .2.
	Subnet string
}

func (s *Service) List() ([]db.Peer, error) {
	return s.DB.ListPeers()
}

type CreateRequest struct {
	Name string `json:"name"`
}

func (s *Service) Create(req CreateRequest) (db.Peer, error) {
	if req.Name == "" {
		return db.Peer{}, fmt.Errorf("name is required")
	}

	priv, pub, err := s.AWG.GenerateKeypair()
	if err != nil {
		return db.Peer{}, err
	}

	addr, err := s.nextFreeAddress()
	if err != nil {
		return db.Peer{}, err
	}

	peer := db.Peer{
		Name:       req.Name,
		PublicKey:  pub,
		PrivateKey: priv,
		AddressV4:  addr,
	}
	id, err := s.DB.InsertPeer(peer)
	if err != nil {
		return db.Peer{}, err
	}
	peer, err = s.DB.GetPeer(id)
	if err != nil {
		return db.Peer{}, err
	}
	peer.PrivateKey = priv // carry in-memory for the immediate config download

	if err := s.syncServerConf(); err != nil {
		return peer, fmt.Errorf("sync awg conf: %w", err)
	}
	return peer, nil
}

func (s *Service) Delete(id int64) error {
	if err := s.DB.DeletePeer(id); err != nil {
		return err
	}
	return s.syncServerConf()
}

// ClientConfig renders the .conf file the user imports into their
// Amnezia client. Includes the private key once — after this moment,
// the backend never returns the private key again.
func (s *Service) ClientConfig(id int64) (string, error) {
	peer, err := s.DB.GetPeer(id)
	if err != nil {
		return "", err
	}
	if peer.PrivateKey == "" {
		return "", fmt.Errorf("private key already delivered; create a new peer to re-enroll")
	}
	data := struct {
		Peer            db.Peer
		Config          *config.Config
		ServerPublicKey string
		Endpoint        string
	}{
		Peer:            peer,
		Config:          s.Config,
		ServerPublicKey: s.Config.PublicKey,
		Endpoint:        fmt.Sprintf("%s:%d", s.Config.AppDomain, s.Config.ListenPort),
	}
	var buf bytes.Buffer
	if err := s.ClientTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *Service) QRCode(id int64) ([]byte, error) {
	conf, err := s.ClientConfig(id)
	if err != nil {
		return nil, err
	}
	return qrcode.Encode(conf, qrcode.Medium, 512)
}

// nextFreeAddress returns 10.9.0.N/32 for the smallest free N ≥ 2.
func (s *Service) nextFreeAddress() (string, error) {
	used, err := s.DB.UsedAddresses()
	if err != nil {
		return "", err
	}
	_, subnet, err := net.ParseCIDR(s.Subnet)
	if err != nil {
		return "", fmt.Errorf("parse subnet %q: %w", s.Subnet, err)
	}
	base := subnet.IP.To4()
	if base == nil {
		return "", fmt.Errorf("only IPv4 subnets supported")
	}
	for i := 2; i < 254; i++ {
		addr := fmt.Sprintf("%d.%d.%d.%d/32", base[0], base[1], base[2], i)
		if !used[addr] {
			return addr, nil
		}
	}
	return "", fmt.Errorf("address pool exhausted")
}
