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

type Service struct {
	DB     *db.DB
	AWG    *awg.Client
	Config *config.Config

	ServerTemplate *template.Template
	ClientTemplate *template.Template

	ServerConfPath string
	AwgQuickBinary string
	Subnet         string
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
	// Carried in-memory only: the private key is delivered exactly once
	// in the create response, then dropped from the DB row.
	peer.PrivateKey = priv

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
