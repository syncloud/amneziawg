package installer

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/syncloud/golib/config"
	"github.com/syncloud/golib/linux"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"

	"hooks/obfuscation"
	"hooks/portpicker"
)

const App = "amneziawg"

// Variables is the template context passed to config/*.tpl files by
// golib/config.Generate.
type Variables struct {
	App              string
	AppDir           string
	DataDir          string
	CommonDir        string
	AppUrl           string
	AppDomain        string
	Domain           string
	Secret           string
	OIDCClientSecret string

	// Server
	ServerPrivateKey string
	ServerPublicKey  string
	ListenPort       int

	// Obfuscation
	Jc   int
	Jmin int
	Jmax int
	S1   int
	S2   int
	H1   uint32
	H2   uint32
	H3   uint32
	H4   uint32

	// Peers (populated by the backend, not the installer).
	Peers []Peer
}

type Peer struct {
	Name       string
	PublicKey  string
	AllowedIPs string
}

type Installer struct {
	newVersionFile     string
	currentVersionFile string
	configDir          string
	platformClient     *platform.Client
	installFile        string
	appDir             string
	dataDir            string
	commonDir          string
	executor           *Executor
	logger             *zap.Logger
}

func New(logger *zap.Logger) *Installer {
	appDir := fmt.Sprintf("/snap/%s/current", App)
	dataDir := fmt.Sprintf("/var/snap/%s/current", App)
	commonDir := fmt.Sprintf("/var/snap/%s/common", App)
	configDir := path.Join(dataDir, "config")

	return &Installer{
		newVersionFile:     path.Join(appDir, "version"),
		currentVersionFile: path.Join(dataDir, "version"),
		configDir:          configDir,
		platformClient:     platform.New(),
		installFile:        path.Join(dataDir, "installed"),
		appDir:             appDir,
		dataDir:            dataDir,
		commonDir:          commonDir,
		executor:           NewExecutor(logger),
		logger:             logger,
	}
}

func (i *Installer) Install() error {
	if err := linux.CreateUser(App); err != nil {
		return err
	}
	if err := i.initServerState(); err != nil {
		return err
	}
	if err := i.UpdateConfigs(); err != nil {
		return err
	}
	if err := i.FixPermissions(); err != nil {
		return err
	}
	return i.StorageChange()
}

func (i *Installer) Configure() error {
	if i.IsInstalled() {
		if err := i.Upgrade(); err != nil {
			return err
		}
	} else {
		if err := i.Initialize(); err != nil {
			return err
		}
	}
	if err := i.FixPermissions(); err != nil {
		return err
	}
	return i.UpdateVersion()
}

func (i *Installer) Initialize() error {
	if err := i.StorageChange(); err != nil {
		return err
	}
	return os.WriteFile(i.installFile, []byte("installed"), 0644)
}

func (i *Installer) Upgrade() error {
	return i.StorageChange()
}

func (i *Installer) IsInstalled() bool {
	_, err := os.Stat(i.installFile)
	return err == nil
}

func (i *Installer) PreRefresh() error {
	return nil
}

func (i *Installer) PostRefresh() error {
	if err := i.UpdateConfigs(); err != nil {
		return err
	}
	if err := i.ClearVersion(); err != nil {
		return err
	}
	return i.FixPermissions()
}

func (i *Installer) StorageChange() error {
	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}
	return linux.Chown(storageDir, App)
}

func (i *Installer) ClearVersion() error {
	return os.RemoveAll(i.currentVersionFile)
}

func (i *Installer) UpdateVersion() error {
	data, err := os.ReadFile(i.newVersionFile)
	if err != nil {
		return err
	}
	return os.WriteFile(i.currentVersionFile, data, 0644)
}

// initServerState runs only on first install. Generates the server
// keypair, obfuscation parameters, and picks the UDP listen port.
// All values are persisted so subsequent configure runs reuse them —
// rotating any of these would force every client to re-enroll.
func (i *Installer) initServerState() error {
	for _, d := range []string{i.dataDir, i.configDir, path.Join(i.dataDir, "nginx"), path.Join(i.commonDir, "db")} {
		if err := linux.CreateMissingDirs(d); err != nil {
			return err
		}
	}

	if _, err := os.Stat(path.Join(i.dataDir, "server.key")); os.IsNotExist(err) {
		priv, pub, err := i.generateServerKeypair()
		if err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(i.dataDir, "server.key"), []byte(priv), 0600); err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(i.dataDir, "server.pub"), []byte(pub), 0644); err != nil {
			return err
		}
	}

	if _, err := os.Stat(path.Join(i.dataDir, "obfuscation.json")); os.IsNotExist(err) {
		p, err := obfuscation.Generate()
		if err != nil {
			return err
		}
		if err := writeJSON(path.Join(i.dataDir, "obfuscation.json"), p); err != nil {
			return err
		}
	}

	if _, err := os.Stat(path.Join(i.dataDir, "port")); os.IsNotExist(err) {
		port, err := portpicker.Pick()
		if err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(i.dataDir, "port"), []byte(fmt.Sprintf("%d", port)), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateServerKeypair shells out to `awg genkey` / `awg pubkey` from
// the bundled amneziawg-tools. We don't reimplement Curve25519 — the
// upstream binary is the source of truth.
func (i *Installer) generateServerKeypair() (string, string, error) {
	awg := path.Join(i.appDir, "amneziawg-tools", "bin", "awg")
	privOut, err := i.executor.Run(awg, "genkey")
	if err != nil {
		return "", "", fmt.Errorf("awg genkey: %w", err)
	}
	priv := strings.TrimSpace(privOut)

	cmd := fmt.Sprintf("echo %s | %s pubkey", priv, awg)
	pubOut, err := i.executor.Run("sh", "-c", cmd)
	if err != nil {
		return "", "", fmt.Errorf("awg pubkey: %w", err)
	}
	pub := strings.TrimSpace(pubOut)
	return priv, pub, nil
}

func (i *Installer) UpdateConfigs() error {
	if err := linux.CreateMissingDirs(
		path.Join(i.dataDir, "nginx"),
		path.Join(i.configDir),
		path.Join(i.commonDir, "db"),
	); err != nil {
		return err
	}
	if err := linux.Chown(i.dataDir, App); err != nil {
		return err
	}

	appUrl, err := i.platformClient.GetAppUrl(App)
	if err != nil {
		return err
	}
	appDomain, err := i.platformClient.GetAppDomainName(App)
	if err != nil {
		return err
	}
	domain, found := strings.CutPrefix(appDomain, App+".")
	if !found {
		return fmt.Errorf("%s is not a prefix of %s", App, appDomain)
	}

	secret, err := getOrCreateUuid(path.Join(i.dataDir, ".secret"))
	if err != nil {
		return err
	}

	// Register as an OIDC client with the platform's Authelia on every
	// configure so redirect URI / client_secret stay in sync. Pattern
	// cribbed from ../paperless/cli/installer/installer.go.
	oidcSecret, err := i.platformClient.RegisterOIDCClient(
		App,
		"/auth/callback",
		true, // require PKCE
		"client_secret_basic",
	)
	if err != nil {
		return fmt.Errorf("register oidc client: %w", err)
	}

	priv, err := os.ReadFile(path.Join(i.dataDir, "server.key"))
	if err != nil {
		return err
	}
	pub, err := os.ReadFile(path.Join(i.dataDir, "server.pub"))
	if err != nil {
		return err
	}
	var obfParams obfuscation.Params
	if err := readJSON(path.Join(i.dataDir, "obfuscation.json"), &obfParams); err != nil {
		return err
	}
	portRaw, err := os.ReadFile(path.Join(i.dataDir, "port"))
	if err != nil {
		return err
	}
	var port int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(portRaw)), "%d", &port); err != nil {
		return err
	}

	variables := Variables{
		App:              App,
		AppDir:           i.appDir,
		DataDir:          i.dataDir,
		CommonDir:        i.commonDir,
		AppUrl:           appUrl,
		AppDomain:        appDomain,
		Domain:           domain,
		Secret:           secret,
		OIDCClientSecret: oidcSecret,
		ServerPrivateKey: strings.TrimSpace(string(priv)),
		ServerPublicKey:  strings.TrimSpace(string(pub)),
		ListenPort:       port,
		Jc:               obfParams.Jc,
		Jmin:             obfParams.Jmin,
		Jmax:             obfParams.Jmax,
		S1:               obfParams.S1,
		S2:               obfParams.S2,
		H1:               obfParams.H1,
		H2:               obfParams.H2,
		H3:               obfParams.H3,
		H4:               obfParams.H4,
	}

	return config.Generate(
		path.Join(i.appDir, "config"),
		i.configDir,
		variables,
	)
}

func (i *Installer) BackupPreStop() error {
	return i.PreRefresh()
}

func (i *Installer) RestorePreStart() error {
	return i.PostRefresh()
}

func (i *Installer) RestorePostStart() error {
	return i.Configure()
}

func (i *Installer) AccessChange() error {
	return i.UpdateConfigs()
}

func (i *Installer) FixPermissions() error {
	if err := linux.Chown(i.dataDir, App); err != nil {
		return err
	}
	return linux.Chown(i.commonDir, App)
}

func getOrCreateUuid(file string) (string, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		secret := uuid.New().String()
		if err := os.WriteFile(file, []byte(secret), 0600); err != nil {
			return "", err
		}
		return secret, nil
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
