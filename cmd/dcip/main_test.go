package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dcip/dcip/network/p2p"
)

func TestLoadConfigCreatesDefaults(t *testing.T) {
	app := &application{
		stdout:       &bytes.Buffer{},
		stderr:       &bytes.Buffer{},
		configPath:   filepath.Join(t.TempDir(), "config.yaml"),
		identityPath: filepath.Join(t.TempDir(), "identity.key"),
		chainPath:    filepath.Join(t.TempDir(), "chain"),
	}

	cfg, err := app.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.Node.Port != p2p.DefaultPort {
		t.Fatalf("cfg.Node.Port = %d, want %d", cfg.Node.Port, p2p.DefaultPort)
	}
	if _, err := os.Stat(app.configPath); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}

func TestWalletNewCommandCreatesIdentity(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := &application{
		stdout:       stdout,
		stderr:       &bytes.Buffer{},
		configPath:   filepath.Join(t.TempDir(), "config.yaml"),
		identityPath: filepath.Join(t.TempDir(), "identity.key"),
		chainPath:    filepath.Join(t.TempDir(), "chain"),
	}

	cmd := app.rootCmd()
	cmd.SetArgs([]string{"wallet", "new"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(app.identityPath); err != nil {
		t.Fatalf("expected identity file to exist: %v", err)
	}
	if !strings.Contains(stdout.String(), "Address: DCIP") {
		t.Fatalf("unexpected wallet output: %q", stdout.String())
	}
}

func TestVersionCommandOutputsVersion(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := &application{
		stdout:       stdout,
		stderr:       &bytes.Buffer{},
		configPath:   filepath.Join(t.TempDir(), "config.yaml"),
		identityPath: filepath.Join(t.TempDir(), "identity.key"),
		chainPath:    filepath.Join(t.TempDir(), "chain"),
	}

	cmd := app.rootCmd()
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "DCIP Node v"+version) {
		t.Fatalf("unexpected version output: %q", stdout.String())
	}
}

func TestQueryCommandUsesEchoAdapter(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("node:\n  port: 7337\n  role: agent\ninference:\n  adapter: echo\nnetwork:\n  bootstrap: []\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := &application{
		stdout:       stdout,
		stderr:       &bytes.Buffer{},
		configPath:   configPath,
		identityPath: filepath.Join(dir, "identity.key"),
		chainPath:    filepath.Join(dir, "chain"),
	}

	cmd := app.rootCmd()
	cmd.SetArgs([]string{"query", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "hello" {
		t.Fatalf("unexpected query output: %q", stdout.String())
	}
}
