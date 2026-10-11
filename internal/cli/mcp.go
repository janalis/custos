package cli

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"custos/internal/editor/mcp"
	"custos/internal/platform/safeio"
	"custos/internal/project/config"
)

func cmdMCP(version string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := newFlagSet("mcp", stderr)
	stdio := fs.Bool("stdio", false, "communicate over stdin/stdout (the default)")
	root := fs.String("root", "", "default workspace root")
	settings := fs.String("workspaces", "", "private JSON workspace, identity and transport configuration")
	storage := fs.String("storage", "", "private checkpoint directory outside workspaces")
	listen := fs.String("http", "", "serve Streamable HTTP at this address (default transport: stdio)")
	cert := fs.String("tls-cert", "", "HTTPS certificate for HTTP transport")
	key := fs.String("tls-key", "", "HTTPS private key for HTTP transport")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *stdio && *listen != "" {
		return errors.New("--stdio and --http are mutually exclusive")
	}
	if len(fs.Args()) > 0 {
		return errors.New("mcp accepts flags only")
	}
	var options mcp.Options
	if *settings != "" {
		data, err := safeio.ReadFile(*settings, config.MaxFileSize)
		if err != nil {
			return err
		}
		options, err = mcp.ParseOptions(data)
		if err != nil {
			return err
		}
		if *root != "" {
			return errors.New("--root cannot be combined with explicit workspaces")
		}
	} else {
		chosen := *root
		provenance := "flag"
		if chosen == "" {
			chosen = os.Getenv("CUSTOS_COMPOSER_ROOT")
			provenance = "composer launcher"
		}
		if chosen == "" {
			cfg, err := config.Load(".")
			if err != nil {
				return err
			}
			chosen = cfg.Root
			provenance = "working directory discovery"
		}
		options.Workspaces = []mcp.Workspace{{ID: "default", Root: chosen, Readers: []string{"local"}, Writers: []string{"local"}, Provenance: provenance}}
		options.Default = "default"
	}
	if *storage != "" {
		options.Storage = *storage
	}
	if *listen != "" {
		host, _, err := net.SplitHostPort(*listen)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			if *cert == "" || *key == "" || !strings.HasPrefix(options.PublicURL, "https://") || len(options.Hosts) == 0 || len(options.Origins) == 0 {
				return errors.New("public HTTP requires TLS, HTTPS publicUrl and explicit hosts/origins")
			}
		}
		if len(options.Hosts) == 0 {
			options.Hosts = []string{*listen}
		}
	}
	server, err := mcp.New(options, registry())
	if err != nil {
		return err
	}
	defer server.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *listen == "" {
		return server.ServeStdio(ctx, version, stdin, stdout)
	}
	handler, err := server.HTTPHandler(version)
	if err != nil {
		return err
	}
	httpServer := &http.Server{Addr: *listen, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 2 * time.Minute, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	return serveMCPHTTP(ctx, httpServer, *cert, *key)
}

func serveMCPHTTP(ctx context.Context, httpServer *http.Server, cert, key string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	var err error
	if cert != "" || key != "" {
		err = httpServer.ListenAndServeTLS(cert, key)
	} else {
		err = httpServer.ListenAndServe()
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
