package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Obfuscation struct {
	Jc   int    `json:"Jc"`
	Jmin int    `json:"Jmin"`
	Jmax int    `json:"Jmax"`
	S1   int    `json:"S1"`
	S2   int    `json:"S2"`
	H1   uint32 `json:"H1"`
	H2   uint32 `json:"H2"`
	H3   uint32 `json:"H3"`
	H4   uint32 `json:"H4"`
}

type Config struct {
	DataDir string

	PrivateKey       string
	PublicKey        string
	ListenPort       int
	AppDomain        string
	AppUrl           string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCAuthBaseURL  string
	OIDCRedirectURI  string
	Obfuscation      Obfuscation
}

func (c *Config) Load() error {
	priv, err := readTrim(filepath.Join(c.DataDir, "server.key"))
	if err != nil {
		return fmt.Errorf("server.key: %w", err)
	}
	c.PrivateKey = priv

	pub, err := readTrim(filepath.Join(c.DataDir, "server.pub"))
	if err != nil {
		return fmt.Errorf("server.pub: %w", err)
	}
	c.PublicKey = pub

	portStr, err := readTrim(filepath.Join(c.DataDir, "port"))
	if err != nil {
		return fmt.Errorf("port: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("parse port %q: %w", portStr, err)
	}
	c.ListenPort = port

	obfData, err := os.ReadFile(filepath.Join(c.DataDir, "obfuscation.json"))
	if err != nil {
		return fmt.Errorf("obfuscation.json: %w", err)
	}
	if err := json.Unmarshal(obfData, &c.Obfuscation); err != nil {
		return fmt.Errorf("parse obfuscation.json: %w", err)
	}

	oidc, err := loadKV(filepath.Join(c.DataDir, "config", "oidc.env"))
	if err != nil {
		return fmt.Errorf("oidc.env: %w", err)
	}
	c.AppDomain = oidc["APP_DOMAIN"]
	c.AppUrl = oidc["APP_URL"]
	c.OIDCClientID = oidc["OIDC_CLIENT_ID"]
	c.OIDCClientSecret = oidc["OIDC_CLIENT_SECRET"]
	c.OIDCAuthBaseURL = oidc["OIDC_AUTH_BASE_URL"]
	c.OIDCRedirectURI = oidc["OIDC_REDIRECT_URI"]

	return nil
}

func readTrim(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func loadKV(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
	}
	return out, nil
}
