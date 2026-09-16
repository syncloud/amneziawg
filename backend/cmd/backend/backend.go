package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"go.uber.org/zap"

	"backend/auth"
	"backend/awg"
	"backend/config"
	"backend/db"
	"backend/peers"
	"backend/status"
)

const (
	app          = "amneziawg"
	appDir       = "/snap/" + app + "/current"
	dataDir      = "/var/snap/" + app + "/current"
	templatesDir = appDir + "/templates"
	backendSock  = dataDir + "/backend.sock"
	dbPath       = dataDir + "/db/data.db"
	secretPath   = dataDir + "/.secret"
	awgBin       = appDir + "/amneziawg-tools/bin/awg"
	awgQuickBin  = appDir + "/amneziawg-tools/bin/awg-quick"
	serverIface  = "awg0"
	serverSubnet = "10.9.0.0/24"
	adminGroup   = "syncloud"
)

var serverConfPath = filepath.Join(dataDir, "config", serverIface+".conf")

type backend struct {
	config *config.Config
	peers  *peers.Service
	status *status.Service
}

func newBackend() (*backend, error) {
	cfg := &config.Config{DataDir: dataDir}
	if err := cfg.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	awgClient := &awg.Client{Binary: awgBin, Interface: serverIface}

	serverTpl, err := template.ParseFiles(filepath.Join(templatesDir, "awg-server.conf.tpl"))
	if err != nil {
		return nil, fmt.Errorf("parse server template: %w", err)
	}
	clientTpl, err := template.ParseFiles(filepath.Join(templatesDir, "awg-client.conf.tpl"))
	if err != nil {
		return nil, fmt.Errorf("parse client template: %w", err)
	}

	return &backend{
		config: cfg,
		peers: &peers.Service{
			DB:             database,
			AWG:            awgClient,
			Config:         cfg,
			ServerTemplate: serverTpl,
			ClientTemplate: clientTpl,
			ServerConfPath: serverConfPath,
			AwgQuickBinary: awgQuickBin,
			Subnet:         serverSubnet,
		},
		status: &status.Service{AWG: awgClient, Config: cfg},
	}, nil
}

func (b *backend) RenderServerConf() error {
	return b.peers.RenderServerConf()
}

func (b *backend) Serve(logger *zap.Logger) error {
	cookieSecret, err := os.ReadFile(secretPath)
	if err != nil {
		return fmt.Errorf("read cookie secret: %w", err)
	}

	oidc := &auth.OIDC{
		IssuerURL:    b.config.OIDCAuthBaseURL,
		ClientID:     b.config.OIDCClientID,
		ClientSecret: b.config.OIDCClientSecret,
		RedirectURL:  b.config.OIDCRedirectURI,
		AdminGroup:   adminGroup,
		CookieSecret: cookieSecret,
		Logger:       logger,
	}
	if err := oidc.Init(context.Background()); err != nil {
		return fmt.Errorf("oidc init: %w", err)
	}

	apiMux := http.NewServeMux()
	b.peers.RegisterRoutes(apiMux)
	b.status.RegisterRoutes(apiMux)

	rootMux := http.NewServeMux()
	rootMux.HandleFunc("GET /auth/login", oidc.Login)
	rootMux.HandleFunc("GET /auth/callback", oidc.Callback)
	rootMux.HandleFunc("GET /auth/logout", oidc.Logout)
	rootMux.Handle("/api/", oidc.Middleware(apiMux))

	_ = os.Remove(backendSock)
	listener, err := net.Listen("unix", backendSock)
	if err != nil {
		return fmt.Errorf("listen %s: %w", backendSock, err)
	}
	if err := os.Chmod(backendSock, 0666); err != nil {
		return fmt.Errorf("chmod socket: %w", err)
	}

	logger.Info("backend listening", zap.String("socket", backendSock))
	return (&http.Server{Handler: rootMux}).Serve(listener)
}
