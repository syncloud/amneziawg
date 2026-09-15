package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
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
	configDir    = appDir + "/config"
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

func main() {
	cmd := &cobra.Command{
		Use:          "backend",
		Short:        "AmneziaWG web backend — peer management API + OIDC auth",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger, err := buildLogger()
			if err != nil {
				return err
			}
			return run(logger)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:          "render-server-conf",
		Short:        "Regenerate " + serverIface + ".conf from the peers database",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			b, err := newBackend()
			if err != nil {
				return err
			}
			return b.peers.RenderServerConf()
		},
	})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type backend struct {
	config *config.Config
	awg    *awg.Client
	peers  *peers.Service
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
		awg:    awgClient,
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
	}, nil
}

func run(logger *zap.Logger) error {
	b, err := newBackend()
	if err != nil {
		return err
	}
	cfg := b.config
	peersService := b.peers

	statusService := &status.Service{AWG: b.awg, Config: cfg}

	cookieSecret, err := os.ReadFile(secretPath)
	if err != nil {
		return fmt.Errorf("read cookie secret: %w", err)
	}

	oidc := &auth.OIDC{
		IssuerURL:    cfg.OIDCAuthBaseURL,
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		RedirectURL:  cfg.OIDCRedirectURI,
		AdminGroup:   adminGroup,
		CookieSecret: cookieSecret,
		Logger:       logger,
	}
	if err := oidc.Init(context.Background()); err != nil {
		return fmt.Errorf("oidc init: %w", err)
	}

	apiMux := http.NewServeMux()
	peersService.RegisterRoutes(apiMux)
	statusService.RegisterRoutes(apiMux)

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

func buildLogger() (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.Encoding = "console"
	c.EncoderConfig.TimeKey = ""
	c.OutputPaths = []string{"stdout"}
	c.ErrorOutputPaths = []string{"stderr"}
	return c.Build()
}
