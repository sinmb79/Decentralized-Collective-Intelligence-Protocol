package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/dcip/dcip/core/chain"
	"github.com/dcip/dcip/core/identity"
	"github.com/dcip/dcip/inference/adapter"
	"github.com/dcip/dcip/network/p2p"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const version = "0.1.0"

type config struct {
	Node      nodeConfig      `yaml:"node"`
	Inference inferenceConfig `yaml:"inference"`
	Network   networkConfig   `yaml:"network"`
}

type nodeConfig struct {
	Port int    `yaml:"port"`
	Role string `yaml:"role"`
}

type inferenceConfig struct {
	Adapter     string `yaml:"adapter"`
	OllamaURL   string `yaml:"ollama_url"`
	OllamaModel string `yaml:"ollama_model"`
	OpenAIURL   string `yaml:"openai_url"`
	OpenAIModel string `yaml:"openai_model"`
}

type networkConfig struct {
	Bootstrap []string `yaml:"bootstrap"`
}

type application struct {
	stdout       io.Writer
	stderr       io.Writer
	configPath   string
	identityPath string
	chainPath    string
}

func main() {
	app := newApplication()
	if err := app.rootCmd().Execute(); err != nil {
		fmt.Fprintln(app.stderr, err)
		os.Exit(1)
	}
}

func newApplication() *application {
	base := defaultDataDir()
	return &application{
		stdout:       os.Stdout,
		stderr:       os.Stderr,
		configPath:   filepath.Join(base, "config.yaml"),
		identityPath: filepath.Join(base, "identity.key"),
		chainPath:    filepath.Join(base, "chain"),
	}
}

func (a *application) rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dcip",
		Short: "DCIP L1 blockchain node",
	}

	cmd.PersistentFlags().StringVar(&a.configPath, "config", a.configPath, "path to config file")
	cmd.AddCommand(a.startCmd())
	cmd.AddCommand(a.walletCmd())
	cmd.AddCommand(a.queryCmd())
	cmd.AddCommand(a.peersCmd())
	cmd.AddCommand(a.versionCmd())
	return cmd
}

func (a *application) startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start a DCIP node",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.loadConfig()
			if err != nil {
				return err
			}

			id, err := a.ensureIdentity()
			if err != nil {
				return err
			}

			backend, err := adapter.New("", cfg.adapterOptions())
			if err != nil {
				return err
			}

			node, err := p2p.NewP2PNode(id, cfg.Node.Port)
			if err != nil {
				return err
			}
			defer node.Close()

			p2p.BootstrapPeers = append([]string(nil), cfg.Network.Bootstrap...)
			startErr := node.Start()

			a.printBanner(id.Address, cfg.Node.Role, cfg.Node.Port, node.PeerCount())
			fmt.Fprintf(a.stdout, "Adapter: %s (%s)\n", cfg.Inference.Adapter, backend.ModelID())
			if startErr != nil {
				fmt.Fprintf(a.stderr, "warning: discovery degraded: %v\n", startErr)
			}
			if !backend.IsReady() {
				fmt.Fprintf(a.stderr, "warning: inference adapter %q is not ready\n", cfg.Inference.Adapter)
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sigCh)

			<-sigCh
			fmt.Fprintln(a.stdout, "Graceful shutdown complete.")
			return nil
		},
	}
}

func (a *application) walletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage the local DCIP wallet",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "new",
		Short: "Create a new wallet",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := identity.NewIdentity()
			if err != nil {
				return err
			}
			if err := id.SaveIdentity(a.identityPath); err != nil {
				return err
			}

			fmt.Fprintf(a.stdout, "Address: %s\n", id.Address)
			fmt.Fprintf(a.stdout, "Saved to: %s\n", a.identityPath)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "balance",
		Short: "Show the local wallet balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := identity.LoadIdentity(a.identityPath)
			if err != nil {
				return err
			}

			chainDB, err := chain.Open(a.chainPath)
			if err != nil {
				return err
			}
			defer chainDB.Close()

			balance := chainDB.State().Balance(id.Address)
			fmt.Fprintf(a.stdout, "ACL Balance: %s\n", formatACL(balance))
			return nil
		},
	})

	return cmd
}

func (a *application) queryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query [text]",
		Short: "Run a local inference query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.loadConfig()
			if err != nil {
				return err
			}

			backend, err := adapter.New("", cfg.adapterOptions())
			if err != nil {
				return err
			}
			if !backend.IsReady() {
				return errors.New("configured inference adapter is not ready")
			}

			response, err := backend.Infer(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(a.stdout, response)
			return nil
		},
	}
}

func (a *application) peersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peers",
		Short: "Show currently known peers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.loadConfig()
			if err != nil {
				return err
			}

			id, err := a.ensureIdentity()
			if err != nil {
				return err
			}

			node, err := p2p.NewP2PNode(id, cfg.Node.Port)
			if err != nil {
				return err
			}
			defer node.Close()

			p2p.BootstrapPeers = append([]string(nil), cfg.Network.Bootstrap...)
			if err := node.Start(); err != nil {
				fmt.Fprintf(a.stderr, "warning: %v\n", err)
			}

			fmt.Fprintf(a.stdout, "Peers: %d\n", node.PeerCount())
			for _, peerID := range node.PeerList() {
				fmt.Fprintln(a.stdout, peerID)
			}
			return nil
		},
	}
}

func (a *application) versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the DCIP node version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(a.stdout, "DCIP Node v%s\n", version)
		},
	}
}

func (a *application) loadConfig() (config, error) {
	path, err := expandPath(a.configPath)
	if err != nil {
		return config{}, err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := defaultConfig()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return config{}, err
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return config{}, err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return config{}, err
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	cfg.normalize()
	return cfg, nil
}

func (a *application) ensureIdentity() (*identity.Identity, error) {
	id, err := identity.LoadIdentity(a.identityPath)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	id, err = identity.NewIdentity()
	if err != nil {
		return nil, err
	}
	if err := id.SaveIdentity(a.identityPath); err != nil {
		return nil, err
	}
	return id, nil
}

func (a *application) printBanner(address, role string, port, peers int) {
	fmt.Fprintln(a.stdout, "========================================")
	fmt.Fprintf(a.stdout, "  DCIP Node v%s\n", version)
	fmt.Fprintf(a.stdout, "  Address: %s\n", address)
	fmt.Fprintf(a.stdout, "  Role:    %s\n", role)
	fmt.Fprintf(a.stdout, "  Port:    %d\n", port)
	fmt.Fprintf(a.stdout, "  Peers:   %d\n", peers)
	fmt.Fprintln(a.stdout, "========================================")
	fmt.Fprintln(a.stdout, "\"Alone we are limited. Together we are intelligence.\"")
}

func defaultConfig() config {
	return config{
		Node: nodeConfig{
			Port: p2p.DefaultPort,
			Role: identity.RoleAgent,
		},
		Inference: inferenceConfig{
			Adapter:     adapter.KindOllama,
			OllamaURL:   adapter.DefaultOllamaURL,
			OllamaModel: adapter.DefaultOllamaModel,
			OpenAIURL:   adapter.DefaultOpenAIURL,
			OpenAIModel: adapter.DefaultOpenAIModel,
		},
		Network: networkConfig{
			Bootstrap: []string{},
		},
	}
}

func (c *config) normalize() {
	if c.Node.Port <= 0 {
		c.Node.Port = p2p.DefaultPort
	}
	if strings.TrimSpace(c.Node.Role) == "" {
		c.Node.Role = identity.RoleAgent
	}
	if strings.TrimSpace(c.Inference.Adapter) == "" {
		c.Inference.Adapter = adapter.KindOllama
	}
	if strings.TrimSpace(c.Inference.OllamaURL) == "" {
		c.Inference.OllamaURL = adapter.DefaultOllamaURL
	}
	if strings.TrimSpace(c.Inference.OllamaModel) == "" {
		c.Inference.OllamaModel = adapter.DefaultOllamaModel
	}
	if strings.TrimSpace(c.Inference.OpenAIURL) == "" {
		c.Inference.OpenAIURL = adapter.DefaultOpenAIURL
	}
	if strings.TrimSpace(c.Inference.OpenAIModel) == "" {
		c.Inference.OpenAIModel = adapter.DefaultOpenAIModel
	}
}

func (c config) adapterOptions() adapter.Options {
	return adapter.Options{
		Kind:        c.Inference.Adapter,
		OllamaURL:   c.Inference.OllamaURL,
		OllamaModel: c.Inference.OllamaModel,
		OpenAIURL:   c.Inference.OpenAIURL,
		OpenAIModel: c.Inference.OpenAIModel,
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".dcip"
	}

	return filepath.Join(home, ".dcip")
}

func expandPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is empty")
	}
	if path == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

func formatACL(value uint64) string {
	whole := value / 100_000_000
	fraction := value % 100_000_000
	return fmt.Sprintf("%d.%08d", whole, fraction)
}
