package cli

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/editor/mcp"
	"custos/internal/project"
)

func TestMCPCommand(t *testing.T) {
	root := t.TempDir()
	storage := filepath.Join(t.TempDir(), "private")
	t.Setenv("CUSTOS_COMPOSER_ROOT", root)
	for _, args := range [][]string{{"--stdio", "--http", "127.0.0.1:8765"}, {"--bad"}, {"file.php"}, {"--workspaces", filepath.Join(root, "missing")}, {"--http", "invalid"}, {"--http", "0.0.0.0:8765"}, {"--root", filepath.Join(root, "missing")}} {
		if err := cmdMCP("test", args, strings.NewReader(""), io.Discard, io.Discard); err == nil {
			t.Fatal(args)
		}
	}
	if code := Run("test", []string{"mcp", "--root", root, "--storage", storage}, strings.NewReader(""), io.Discard, io.Discard); code != 0 {
		t.Fatal(code)
	}
	if err := cmdMCP("test", []string{"--stdio", "--storage", storage}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUSTOS_COMPOSER_ROOT", "")
	t.Chdir(root)
	if err := cmdMCP("test", []string{"--storage", storage}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "custos.json"), []byte("bad json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := cmdMCP("test", []string{"--storage", storage}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("bad discovered config")
	}
	if err := os.Remove(filepath.Join(root, "custos.json")); err != nil {
		t.Fatal(err)
	}
	options := mcp.Options{Storage: storage, Default: "app", Workspaces: []mcp.Workspace{{ID: "app", Root: root, Readers: []string{"local"}}}}
	configPath := filepath.Join(t.TempDir(), "workspaces.json")
	write := func(value any) {
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(configPath, b, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(options)
	if err := cmdMCP("test", []string{"--workspaces", configPath, "--root", root}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("conflicting root")
	}
	if err := cmdMCP("test", []string{"--workspaces", configPath}, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if err := cmdMCP("test", []string{"--workspaces", configPath, "--http", "127.0.0.1:8765"}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("anonymous HTTP")
	}
	write(map[string]any{"unknown": true})
	if err := cmdMCP("test", []string{"--workspaces", configPath}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("unknown setting")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	options.Credentials = []mcp.Credential{{SHA256: project.Fingerprint([]byte("token")), Identity: "local", Expires: time.Now().Add(time.Hour), Scopes: []string{"custos:read"}}}
	write(options)
	if err := cmdMCP("test", []string{"--workspaces", configPath, "--http", listener.Addr().String()}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("occupied listen unexpectedly succeeded")
	}
	if err := cmdMCP("test", []string{"--workspaces", configPath, "--http", listener.Addr().String(), "--tls-cert", "missing", "--tls-key", "missing"}, strings.NewReader(""), io.Discard, io.Discard); err == nil {
		t.Fatal("invalid TLS unexpectedly succeeded")
	}
}

func TestMCPHTTPShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := serveMCPHTTP(ctx, &http.Server{Addr: "127.0.0.1:0", Handler: http.NotFoundHandler()}, "", "")
	if err != nil {
		t.Fatal(err)
	}
}
